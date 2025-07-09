package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

// ImportOptions options pour l'import CSV
type ImportOptions struct {
	DryRun  bool
	Verbose bool
}

// ImportResult résultats de l'import
type ImportResult struct {
	NewTasks     int
	UpdatedTasks int
	SkippedTasks int
	Errors       []string
	Warnings     []string
}

// ImportCSV importe des tâches depuis un fichier CSV
func (tm *TodoManager) ImportCSV(filename string, mode string, conflict string, options ImportOptions) (*ImportResult, error) {
	result := &ImportResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	if options.Verbose {
		fmt.Printf("📥 Import CSV: %s (mode: %s, conflit: %s)\n", filename, mode, conflict)
	}

	// Vérifier l'existence du fichier
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("fichier CSV introuvable: %s", filename)
	}

	// Ouvrir et parser le CSV
	tasks, parseErrors := tm.parseCSV(filename)
	result.Errors = append(result.Errors, parseErrors...)

	if len(tasks) == 0 {
		return result, fmt.Errorf("aucune tâche valide trouvée dans le CSV")
	}

	if options.Verbose {
		fmt.Printf("📊 %d tâches trouvées dans le CSV\n", len(tasks))
	}

	// Mode replace: supprimer toutes les tâches existantes
	if mode == "replace" {
		if !options.DryRun {
			if !tm.confirmReplace() {
				return result, fmt.Errorf("import annulé par l'utilisateur")
			}
			tm.Tasks = []Task{}
			tm.NextID = 1
		}
		if options.Verbose {
			fmt.Println("🗑️ Toutes les tâches existantes supprimées")
		}
	}

	// Créer une map des tâches existantes par UUID
	existingTasks := make(map[string]*Task)
	for i := range tm.Tasks {
		if tm.Tasks[i].UUID != "" {
			existingTasks[tm.Tasks[i].UUID] = &tm.Tasks[i]
		}
	}

	// Traiter chaque tâche du CSV
	for _, csvTask := range tasks {
		if tm.processImportTask(csvTask, existingTasks, conflict, options, result) {
			if !options.DryRun {
				tm.Tasks = append(tm.Tasks, csvTask)
				if csvTask.ID >= tm.NextID {
					tm.NextID = csvTask.ID + 1
				}
			}
		}
	}

	// Sauvegarder si pas en mode dry-run
	if !options.DryRun && (result.NewTasks > 0 || result.UpdatedTasks > 0) {
		tm.save()
	}

	// Afficher le rapport
	tm.printImportReport(filename, result, options)

	return result, nil
}

// parseCSV parse le fichier CSV et retourne les tâches
func (tm *TodoManager) parseCSV(filename string) ([]Task, []string) {
	var tasks []Task
	var errors []string

	file, err := os.Open(filename)
	if err != nil {
		return tasks, []string{fmt.Sprintf("impossible d'ouvrir le fichier: %v", err)}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Permettre un nombre variable de colonnes

	// Lire l'en-tête
	headers, err := reader.Read()
	if err != nil {
		return tasks, []string{fmt.Sprintf("impossible de lire l'en-tête CSV: %v", err)}
	}

	// Créer une map des colonnes
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[strings.TrimSpace(strings.ToLower(header))] = i
	}

	// Vérifier que la colonne Text existe
	textCol, hasText := columnMap["text"]
	if !hasText {
		return tasks, []string{"colonne 'Text' obligatoire manquante dans le CSV"}
	}

	lineNumber := 1
	for {
		lineNumber++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, fmt.Sprintf("ligne %d: erreur de parsing CSV: %v", lineNumber, err))
			continue
		}

		// Ignorer les lignes vides
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		// Créer la tâche depuis la ligne CSV
		task, lineErrors := tm.createTaskFromCSVRecord(record, columnMap, textCol, lineNumber)
		if len(lineErrors) > 0 {
			errors = append(errors, lineErrors...)
			continue
		}

		if task.Text != "" {
			tasks = append(tasks, task)
		}
	}

	return tasks, errors
}

// createTaskFromCSVRecord crée une tâche depuis un enregistrement CSV
func (tm *TodoManager) createTaskFromCSVRecord(record []string, columnMap map[string]int, textCol int, lineNumber int) (Task, []string) {
	var errors []string

	// Fonction helper pour récupérer une valeur de colonne
	getValue := func(colName string) string {
		if col, exists := columnMap[colName]; exists && col < len(record) {
			return strings.TrimSpace(record[col])
		}
		return ""
	}

	task := Task{
		ID:      tm.NextID,
		Text:    strings.TrimSpace(record[textCol]),
		Done:    false,
		Created: time.Now().Format("2006-01-02 15:04:05"),
		Updated: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Valider que le texte n'est pas vide
	if task.Text == "" {
		errors = append(errors, fmt.Sprintf("ligne %d: texte vide, tâche ignorée", lineNumber))
		return task, errors
	}

	// UUID
	uuidValue := getValue("uuid")
	if uuidValue != "" {
		if tm.isValidUUID(uuidValue) {
			task.UUID = uuidValue
		} else {
			errors = append(errors, fmt.Sprintf("ligne %d: UUID invalide '%s', nouveau UUID généré", lineNumber, uuidValue))
			task.UUID = generateUUID()
		}
	} else {
		task.UUID = generateUUID()
	}

	// Done
	doneValue := strings.ToLower(getValue("done"))
	if doneValue == "true" || doneValue == "1" {
		task.Done = true
	}

	// Priority
	priority := strings.ToLower(getValue("priority"))
	if priority != "" {
		if priority == "high" || priority == "medium" || priority == "low" {
			task.Priority = priority
		} else {
			errors = append(errors, fmt.Sprintf("ligne %d: priorité '%s' invalide, ignorée", lineNumber, priority))
		}
	}

	// Due date
	dueValue := getValue("due")
	if dueValue != "" {
		if validateDate(dueValue) {
			task.Due = dueValue
		} else {
			errors = append(errors, fmt.Sprintf("ligne %d: date '%s' invalide, ignorée", lineNumber, dueValue))
		}
	}

	// Tags
	tagsValue := getValue("tags")
	if tagsValue != "" {
		task.Tags = tm.parseTags(tagsValue)
	}

	// Created date
	createdValue := getValue("created")
	if createdValue != "" {
		if tm.isValidDateTime(createdValue) {
			task.Created = createdValue
		} else {
			errors = append(errors, fmt.Sprintf("ligne %d: date de création '%s' invalide, date actuelle utilisée", lineNumber, createdValue))
		}
	}

	// Updated date
	updatedValue := getValue("updated")
	if updatedValue != "" {
		if tm.isValidDateTime(updatedValue) {
			task.Updated = updatedValue
		} else {
			errors = append(errors, fmt.Sprintf("ligne %d: date de mise à jour '%s' invalide, date actuelle utilisée", lineNumber, updatedValue))
		}
	}

	return task, errors
}

// processImportTask traite une tâche du CSV selon la stratégie de conflit
func (tm *TodoManager) processImportTask(csvTask Task, existingTasks map[string]*Task, conflict string, options ImportOptions, result *ImportResult) bool {
	existingTask, exists := existingTasks[csvTask.UUID]

	if !exists {
		// Nouvelle tâche
		csvTask.ID = tm.NextID
		tm.NextID++
		result.NewTasks++
		if options.Verbose {
			fmt.Printf("➕ Nouvelle tâche: %s\n", csvTask.Text)
		}
		return true
	}

	// Conflit détecté
	if options.Verbose {
		fmt.Printf("🔄 Conflit détecté pour: %s\n", csvTask.Text)
	}

	switch conflict {
	case "skip":
		result.SkippedTasks++
		if options.Verbose {
			fmt.Printf("⏭️ Tâche ignorée (UUID existe déjà)\n")
		}
		return false

	case "update":
		tm.updateExistingTask(existingTask, csvTask)
		result.UpdatedTasks++
		if options.Verbose {
			fmt.Printf("🔄 Tâche mise à jour\n")
		}
		return false

	case "newer":
		if tm.isNewer(csvTask.Updated, existingTask.Updated) {
			tm.updateExistingTask(existingTask, csvTask)
			result.UpdatedTasks++
			if options.Verbose {
				fmt.Printf("🔄 Tâche mise à jour (version plus récente)\n")
			}
		} else {
			result.SkippedTasks++
			if options.Verbose {
				fmt.Printf("⏭️ Tâche ignorée (version plus ancienne)\n")
			}
		}
		return false

	default:
		result.SkippedTasks++
		result.Warnings = append(result.Warnings, fmt.Sprintf("stratégie de conflit '%s' inconnue, tâche ignorée", conflict))
		return false
	}
}

// updateExistingTask met à jour une tâche existante avec les données du CSV
func (tm *TodoManager) updateExistingTask(existing *Task, csvTask Task) {
	existing.Text = csvTask.Text
	existing.Done = csvTask.Done
	existing.Priority = csvTask.Priority
	existing.Due = csvTask.Due
	existing.Tags = csvTask.Tags
	existing.Updated = time.Now().Format("2006-01-02 15:04:05")
}

// confirmReplace demande confirmation pour le mode replace
func (tm *TodoManager) confirmReplace() bool {
	fmt.Printf("⚠️ Mode 'replace': Cela supprimera toutes les %d tâches existantes.\n", len(tm.Tasks))
	fmt.Print("Êtes-vous sûr de vouloir continuer ? (oui/non): ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "oui" || response == "o" || response == "yes" || response == "y"
}

// printImportReport affiche le rapport d'import
func (tm *TodoManager) printImportReport(filename string, result *ImportResult, options ImportOptions) {
	fmt.Printf("\n📥 Import terminé: %s\n", filename)

	if result.NewTasks > 0 {
		fmt.Printf("✅ %d nouvelles tâches\n", result.NewTasks)
	}
	if result.UpdatedTasks > 0 {
		fmt.Printf("🔄 %d tâches mises à jour\n", result.UpdatedTasks)
	}
	if result.SkippedTasks > 0 {
		fmt.Printf("⏭️ %d tâches ignorées\n", result.SkippedTasks)
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️ %d avertissement(s):\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ %d erreur(s):\n", len(result.Errors))
		for _, error := range result.Errors {
			fmt.Printf("  - %s\n", error)
		}
	}

	total := result.NewTasks + result.UpdatedTasks + result.SkippedTasks
	if total > 0 {
		fmt.Printf("\n📊 Total traité: %d tâches\n", total)
	}

	if options.DryRun {
		fmt.Println("\n🔍 Mode dry-run: Aucune modification effectuée")
	}
}

// Fonctions utilitaires

// isValidUUID vérifie si un UUID est valide
func (tm *TodoManager) isValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// isValidDateTime vérifie si une date/heure est valide
func (tm *TodoManager) isValidDateTime(dateTime string) bool {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateTime); err == nil {
			return true
		}
	}
	return false
}

// parseTags parse une chaîne de tags séparés par des espaces
func (tm *TodoManager) parseTags(tagsStr string) []string {
	var tags []string
	words := strings.Fields(tagsStr)

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" && (strings.HasPrefix(word, "+") || strings.HasPrefix(word, "@")) {
			tags = append(tags, word)
		}
	}

	return tags
}

// isNewer compare deux dates et retourne true si la première est plus récente
func (tm *TodoManager) isNewer(date1, date2 string) bool {
	time1, err1 := time.Parse("2006-01-02 15:04:05", date1)
	time2, err2 := time.Parse("2006-01-02 15:04:05", date2)

	if err1 != nil || err2 != nil {
		return false
	}

	return time1.After(time2)
}

// AddImportCommand ajoute la commande import au main
func AddImportCommand() {
	// Cette fonction sera intégrée dans le switch case du main()
	/*
	case "import":
		if len(os.Args) < 3 {
			fmt.Println("❌ Usage: todo import <fichier.csv> [--mode=merge|replace] [--conflict=skip|update|newer] [--dry-run] [--verbose]")
			os.Exit(1)
		}

		filename := os.Args[2]

		// Parse des flags
		importFlags := flag.NewFlagSet("import", flag.ExitOnError)
		mode := importFlags.String("mode", "merge", "Mode d'import (merge, replace)")
		conflict := importFlags.String("conflict", "skip", "Stratégie de conflit (skip, update, newer)")
		dryRun := importFlags.Bool("dry-run", false, "Aperçu sans modification")
		verbose := importFlags.Bool("verbose", false, "Mode verbeux")

		importFlags.Parse(os.Args[3:])

		// Valider les paramètres
		if *mode != "merge" && *mode != "replace" {
			fmt.Println("❌ Mode invalide. Utilisez 'merge' ou 'replace'")
			os.Exit(1)
		}

		if *conflict != "skip" && *conflict != "update" && *conflict != "newer" {
			fmt.Println("❌ Stratégie de conflit invalide. Utilisez 'skip', 'update' ou 'newer'")
			os.Exit(1)
		}

		options := ImportOptions{
			DryRun:  *dryRun,
			Verbose: *verbose,
		}

		_, err := tm.ImportCSV(filename, *mode, *conflict, options)
		if err != nil {
			fmt.Printf("❌ Erreur lors de l'import : %v\n", err)
			os.Exit(1)
		}
	*/
}
