package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	version   = "dev version"     // Version par défaut
	buildTime = "unknown" // Date de build
	gitCommit = "unknown" // Hash du commit
)

// Task représente une tâche
type Task struct {
	ID       int      `json:"id"`
	UUID     string   `json:"uuid"`
	Text     string   `json:"text"`
	Done     bool     `json:"done"`
	Priority string   `json:"priority"`
	Due      string   `json:"due"`
	Tags     []string `json:"tags"`
	Created  string   `json:"created"`
	Updated  string   `json:"updated"`
}

// TodoManager gère les tâches
type TodoManager struct {
	Tasks    []Task `json:"tasks"`
	NextID   int    `json:"nextId"`
	filename string
}

// generateUUID génère un UUID simple (version 4)
func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)

	// Version 4 UUID
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

// Constantes pour les couleurs
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[37m"
	ColorBold   = "\033[1m"
)

// NewTodoManager crée un nouveau gestionnaire de tâches
func NewTodoManager() *TodoManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback sur le répertoire courant
		homeDir = "."
	}

	todoDir := filepath.Join(homeDir, ".todo")
	os.MkdirAll(todoDir, 0755)

	filename := filepath.Join(todoDir, "todo.json")

	tm := &TodoManager{
		Tasks:    []Task{},
		NextID:   1,
		filename: filename,
	}

	tm.load()
	return tm
}

// load charge les tâches depuis le fichier JSON
func (tm *TodoManager) load() {
	data, err := ioutil.ReadFile(tm.filename)
	if err != nil {
		return // Fichier n'existe pas encore
	}

	json.Unmarshal(data, tm)
}

// save sauvegarde les tâches dans le fichier JSON
func (tm *TodoManager) save() error {
	data, err := json.MarshalIndent(tm, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(tm.filename, data, 0644)
}

// Add ajoute une nouvelle tâche avec tags séparés
func (tm *TodoManager) Add(text string, tags []string, priority string, due string) {
	task := Task{
		ID:       tm.NextID,
		UUID:     generateUUID(),
		Text:     text, // Texte intact, AUCUN nettoyage
		Done:     false,
		Priority: priority,
		Due:      due,
		Tags:     tags, // Tags passés en arguments uniquement
		Created:  time.Now().Format("2006-01-02 15:04:05"),
		Updated:  time.Now().Format("2006-01-02 15:04:05"),
	}

	tm.Tasks = append(tm.Tasks, task)
	tm.NextID++
	task.Updated = time.Now().Format("2006-01-02 15:04:05")
	tm.save()

	// Debug - afficher ce qui est sauvegardé
	fmt.Printf("✅ Tâche ajoutée : [%d] %s\n", task.ID, task.Text)
	fmt.Printf("   UUID: %s\n", task.UUID)
	fmt.Printf("   Tags: %v\n", task.Tags)
	fmt.Printf("   Priority: %s\n", task.Priority)
}

// List affiche les tâches
func (tm *TodoManager) List(showDone bool, projectFilter string, contextFilter string, priorityFilter string) {
	filteredTasks := tm.filterTasks(showDone, projectFilter, contextFilter, priorityFilter)

	if len(filteredTasks) == 0 {
		fmt.Println("📝 Aucune tâche trouvée")
		return
	}

	// Trier par priorité puis par date de création
	sort.Slice(filteredTasks, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1, "": 0}
		if priorityOrder[filteredTasks[i].Priority] != priorityOrder[filteredTasks[j].Priority] {
			return priorityOrder[filteredTasks[i].Priority] > priorityOrder[filteredTasks[j].Priority]
		}
		return filteredTasks[i].ID < filteredTasks[j].ID
	})

	for _, task := range filteredTasks {
		tm.printTask(task)
	}
}

// filterTasks filtre les tâches selon les critères
func (tm *TodoManager) filterTasks(showDone bool, projectFilter string, contextFilter string, priorityFilter string) []Task {
	var filtered []Task

	for _, task := range tm.Tasks {
		// Filtre par statut
		if !showDone && task.Done {
			continue
		}

		// Filtre par projet (+tag)
		if projectFilter != "" {
			hasProject := false
			for _, tag := range task.Tags {
				if strings.HasPrefix(tag, "+") && strings.Contains(strings.ToLower(tag), strings.ToLower(projectFilter)) {
					hasProject = true
					break
				}
			}
			if !hasProject {
				continue
			}
		}

		// Filtre par contexte (@tag)
		if contextFilter != "" {
			hasContext := false
			for _, tag := range task.Tags {
				if strings.HasPrefix(tag, "@") && strings.Contains(strings.ToLower(tag), strings.ToLower(contextFilter)) {
					hasContext = true
					break
				}
			}
			if !hasContext {
				continue
			}
		}

		// Filtre par priorité
		if priorityFilter != "" && task.Priority != priorityFilter {
			continue
		}

		filtered = append(filtered, task)
	}

	return filtered
}

// printTask affiche une tâche formatée
func (tm *TodoManager) printTask(task Task) {
	status := "⭕"
	color := ColorReset

	if task.Done {
		status = "✅"
		color = ColorGray
	}

	// Icône de priorité
	priorityIcon := ""
	switch task.Priority {
	case "high":
		priorityIcon = ColorRed + "❗" + ColorReset
	case "medium":
		priorityIcon = ColorYellow + "⚠️" + ColorReset
	case "low":
		priorityIcon = ColorBlue + "ℹ️" + ColorReset
	}

	// Date limite
	dueStr := ""
	if task.Due != "" {
		dueDate, err := time.Parse("2006-01-02", task.Due)
		if err == nil {
			now := time.Now()
			if dueDate.Before(now) {
				dueStr = ColorRed + "[due:" + task.Due + "]" + ColorReset
			} else {
				dueStr = ColorYellow + "[due:" + task.Due + "]" + ColorReset
			}
		}
	}

	// Tags
	tagStr := ""
	if len(task.Tags) > 0 {
		tagStr = " " + ColorBlue + strings.Join(task.Tags, " ") + ColorReset
	}

	// Date de completion
	completedStr := ""
	if task.Done {
		completedStr = " " + ColorGray + "[done:" + task.Updated + "]" + ColorReset
	}

	fmt.Printf("%s[%d] %s %s %s %s%s%s\n",
		color, task.ID, status, priorityIcon, dueStr, task.Text, tagStr, completedStr)
}

// Done marque une tâche comme terminée
func (tm *TodoManager) Done(id int) {
	for i, task := range tm.Tasks {
		if task.ID == id {
			tm.Tasks[i].Done = true
			tm.Tasks[i].Updated = time.Now().Format("2006-01-02 15:04:05")
			tm.save()
			fmt.Printf("✅ Tâche [%d] marquée comme terminée\n", id)
			return
		}
	}
	fmt.Printf("❌ Tâche [%d] introuvable\n", id)
}

// Remove supprime une tâche
func (tm *TodoManager) Remove(id int) {
	for i, task := range tm.Tasks {
		if task.ID == id {
			tm.Tasks = append(tm.Tasks[:i], tm.Tasks[i+1:]...)
			tm.save()
			fmt.Printf("🗑️ Tâche [%d] supprimée\n", id)
			return
		}
	}
	fmt.Printf("❌ Tâche [%d] introuvable\n", id)
}

// Edit modifie une tâche
func (tm *TodoManager) Edit(id int, newText string, tags []string) {
	for i, task := range tm.Tasks {
		if task.ID == id {
			tm.Tasks[i].Text = newText
			tm.Tasks[i].Tags = tags
			tm.Tasks[i].Updated = time.Now().Format("2006-01-02 15:04:05")
			tm.save()
			fmt.Printf("✏️ Tâche [%d] modifiée\n", id)
			return
		}
	}
	fmt.Printf("❌ Tâche [%d] introuvable\n", id)
}

// ExportCSV exporte les tâches en CSV
func (tm *TodoManager) ExportCSV(filename string) error {
	var lines []string
	lines = append(lines, "ID,UUID,Text,Done,Priority,Due,Tags,Created,Updated")

	for _, task := range tm.Tasks {
		line := fmt.Sprintf("%d,%s,\"%s\",%t,%s,%s,\"%s\",%s,%s",
			task.ID,
			task.UUID,
			strings.ReplaceAll(task.Text, "\"", "\"\""),
			task.Done,
			task.Priority,
			task.Due,
			strings.Join(task.Tags, " "),
			task.Created,
			task.Updated,
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return ioutil.WriteFile(filename, []byte(content), 0644)
}

// parsePriority convertit les alias de priorité
func parsePriority(priority string) string {
	switch strings.ToLower(priority) {
	case "h", "high", "haute":
		return "high"
	case "m", "medium", "moyenne":
		return "medium"
	case "l", "low", "basse":
		return "low"
	default:
		return ""
	}
}

// validateDate valide le format de date
func validateDate(dateStr string) bool {
	if dateStr == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// Usage affiche l'aide
func Usage() {
	fmt.Printf("Todo CLI Go %s\n", version)
	fmt.Printf("Build time: %s\n", buildTime)
	fmt.Printf("Git commit: %s\n", gitCommit)

	fmt.Println(`📋 Todo Manager CLI

Usage:
  todo add "Ma tâche" [+projet] [@contexte] [--priority=high] [--due=2025-07-20]
  todo list [--all] [--project=dev] [--context=maison] [--priority=high]
  todo done <id>
  todo remove <id>
  todo edit <id> "Nouveau texte" [+projet] [@contexte]
  todo export [filename.csv]
  todo import <fichier.csv> [--mode=merge|replace] [--conflict=skip|update|newer] [--dry-run] [--verbose]

Options pour add:
  --priority, -p    Priorité (low, medium, high)
  --due, -d        Date limite (format: YYYY-MM-DD)

Options pour list:
  --all, -a        Afficher toutes les tâches (y compris terminées)
  --project       Filtrer par projet (cherche dans les tags +projet)
  --context       Filtrer par contexte (cherche dans les tags @contexte)
  --priority      Filtrer par priorité
  --help, -h      Afficher cette aide

Tags (arguments séparés du texte):
  +projet         Tag de projet (ex: +dev, +travail, +perso)
  @contexte       Tag de contexte/lieu (ex: @maison, @bureau)

Options pour import:
  --mode              Mode d'import (merge, replace) - défaut: merge
  --conflict          Stratégie de conflit (skip, update, newer) - défaut: skip
  --dry-run           Aperçu sans modification
  --verbose           Mode verbeux avec détails

Exemples d'import:
  todo import backup.csv
  todo import tasks.csv --mode=merge --conflict=newer
  todo import external.csv --dry-run --verbose
  todo import full_backup.csv --mode=replace

Exemples:
  todo add "Préparer CV pour xxx@gmail.com" +job @maison --priority=high --due=2025-07-15
  todo add "Calculer 2+2=4" +math @école
  todo add "Email avec +info @dans le texte" +vraitag @vraicontexte
  todo list --project=job
  todo list --context=maison
  todo list --project=job --context=bureau --priority=high
  todo done 1
  todo remove 2
  todo edit 3 "Nouvelle description" +urgent @bureau

Note: Les tags dans le texte ne sont PAS interprétés.
Seuls les arguments +tag @tag après le texte sont utilisés comme tags.`)
}

func main() {
	if len(os.Args) < 2 {
		Usage()
		os.Exit(1)
	}

	tm := NewTodoManager()
	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("❌ Usage: todo add \"Ma tâche\" [+projet] [@contexte] [--priority=high] [--due=2025-07-20]")
			os.Exit(1)
		}

		text := os.Args[2]

		// Extraire les tags des arguments restants (avant les flags)
		var tags []string
		var flagStart = 3

		// Parcourir les arguments pour trouver les tags et où commencent les flags
		for i := 3; i < len(os.Args); i++ {
			arg := os.Args[i]
			if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
				flagStart = i
				break
			}
			// Si l'argument commence par + ou @, c'est un tag
			if strings.HasPrefix(arg, "+") || strings.HasPrefix(arg, "@") {
				tags = append(tags, arg)
			}
		}

		// Parse des flags à partir de flagStart
		addFlags := flag.NewFlagSet("add", flag.ExitOnError)
		priority := addFlags.String("priority", "", "Priorité (low, medium, high)")
		priorityShort := addFlags.String("p", "", "Priorité (alias)")
		due := addFlags.String("due", "", "Date limite (YYYY-MM-DD)")
		dueShort := addFlags.String("d", "", "Date limite (alias)")

		if flagStart < len(os.Args) {
			addFlags.Parse(os.Args[flagStart:])
		} else {
			addFlags.Parse([]string{})
		}

		// Utiliser les alias si les flags principaux sont vides
		if *priority == "" && *priorityShort != "" {
			*priority = *priorityShort
		}
		if *due == "" && *dueShort != "" {
			*due = *dueShort
		}

		*priority = parsePriority(*priority)

		if !validateDate(*due) {
			fmt.Println("❌ Format de date invalide. Utilisez YYYY-MM-DD")
			os.Exit(1)
		}

		tm.Add(text, tags, *priority, *due)

	case "list":
		listFlags := flag.NewFlagSet("list", flag.ExitOnError)
		showAll := listFlags.Bool("all", false, "Afficher toutes les tâches")
		showAllShort := listFlags.Bool("a", false, "Afficher toutes les tâches (alias)")
		project := listFlags.String("project", "", "Filtrer par projet (+tag)")
		context := listFlags.String("context", "", "Filtrer par contexte (@tag)")
		priority := listFlags.String("priority", "", "Filtrer par priorité")

		listFlags.Parse(os.Args[2:])

		showDone := *showAll || *showAllShort
		tm.List(showDone, *project, *context, *priority)

	case "done":
		if len(os.Args) < 3 {
			fmt.Println("❌ Usage: todo done <id>")
			os.Exit(1)
		}

		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("❌ ID invalide")
			os.Exit(1)
		}

		tm.Done(id)

	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("❌ Usage: todo remove <id>")
			os.Exit(1)
		}

		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("❌ ID invalide")
			os.Exit(1)
		}

		tm.Remove(id)

	case "edit":
		if len(os.Args) < 4 {
			fmt.Println("❌ Usage: todo edit <id> \"Nouveau texte\" [+projet] [@contexte]")
			os.Exit(1)
		}

		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("❌ ID invalide")
			os.Exit(1)
		}

		newText := os.Args[3]

		// Extraire les tags des arguments restants (comme pour add)
		var tags []string
		for i := 4; i < len(os.Args); i++ {
			arg := os.Args[i]
			// Si l'argument commence par + ou @, c'est un tag
			if strings.HasPrefix(arg, "+") || strings.HasPrefix(arg, "@") {
				tags = append(tags, arg)
			}
		}

		tm.Edit(id, newText, tags)

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

	case "export":
		filename := "todo_export.csv"
		if len(os.Args) > 2 {
			filename = os.Args[2]
		}

		err := tm.ExportCSV(filename)
		if err != nil {
			fmt.Printf("❌ Erreur lors de l'export : %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("📄 Export terminé : %s\n", filename)
	// Dans le switch case des commandes
	case "version":
		fmt.Printf("Todo CLI Go v%s\n", version)
		fmt.Printf("Build time: %s\n", buildTime)
		fmt.Printf("Git commit: %s\n", gitCommit)
		return

	case "help", "-h", "--help":
		Usage()

	default:
		fmt.Printf("❌ Commande inconnue : %s\n", command)

		Usage()
		os.Exit(1)
	}
}
