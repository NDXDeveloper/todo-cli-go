// todo_manager_test.go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Test helpers
func setupTestEnvironment(t *testing.T) (*TodoManager, string, func()) {
	// Créer un répertoire temporaire pour les tests
	tempDir, err := ioutil.TempDir("", "todo_test")
	if err != nil {
		t.Fatalf("Impossible de créer le répertoire temporaire: %v", err)
	}

	// Créer un TodoManager avec un fichier temporaire
	filename := filepath.Join(tempDir, "test_todo.json")
	tm := &TodoManager{
		Tasks:    []Task{},
		NextID:   1,
		filename: filename,
	}

	// Fonction de nettoyage
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tm, tempDir, cleanup
}

func createSampleTasks() []Task {
	return []Task{
		{
			ID:       1,
			UUID:     "test-uuid-1",
			Text:     "Tâche de test 1",
			Done:     false,
			Priority: "high",
			Due:      "2025-07-20",
			Tags:     []string{"+dev", "@bureau"},
			Created:  "2025-07-09 10:00:00",
			Updated:  "2025-07-09 10:00:00",
		},
		{
			ID:       2,
			UUID:     "test-uuid-2",
			Text:     "Tâche de test 2",
			Done:     true,
			Priority: "medium",
			Tags:     []string{"+perso", "@maison"},
			Created:  "2025-07-08 15:30:00",
			Updated:  "2025-07-09 09:00:00",
		},
	}
}

func assertTaskCount(t *testing.T, tm *TodoManager, expected int) {
	t.Helper()
	if len(tm.Tasks) != expected {
		t.Errorf("Nombre de tâches attendu: %d, obtenu: %d", expected, len(tm.Tasks))
	}
}

func assertTaskExists(t *testing.T, tm *TodoManager, id int) *Task {
	t.Helper()
	for _, task := range tm.Tasks {
		if task.ID == id {
			return &task
		}
	}
	t.Errorf("Tâche avec ID %d introuvable", id)
	return nil
}

// Tests unitaires du TodoManager

func TestNewTodoManager(t *testing.T) {
	t.Run("création avec répertoire valide", func(t *testing.T) {
		tm, _, cleanup := setupTestEnvironment(t)
		defer cleanup()

		if tm == nil {
			t.Fatal("TodoManager ne doit pas être nil")
		}
		if tm.NextID != 1 {
			t.Errorf("NextID initial attendu: 1, obtenu: %d", tm.NextID)
		}
		if len(tm.Tasks) != 0 {
			t.Errorf("Liste initiale doit être vide, obtenu: %d tâches", len(tm.Tasks))
		}
	})
}

func TestTodoManager_Add(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		tags     []string
		priority string
		due      string
		wantErr  bool
	}{
		{
			name:     "tâche simple",
			text:     "Ma tâche de test",
			tags:     []string{},
			priority: "",
			due:      "",
			wantErr:  false,
		},
		{
			name:     "tâche avec tags et priorité",
			text:     "Tâche complexe",
			tags:     []string{"+dev", "@bureau"},
			priority: "high",
			due:      "2025-07-20",
			wantErr:  false,
		},
		{
			name:     "tâche avec priorité invalide",
			text:     "Test priorité",
			tags:     []string{},
			priority: "invalid",
			due:      "",
			wantErr:  false, // Priority invalide devient vide
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, _, cleanup := setupTestEnvironment(t)
			defer cleanup()

			initialCount := len(tm.Tasks)
			tm.Add(tt.text, tt.tags, tt.priority, tt.due)

			// Vérifier qu'une tâche a été ajoutée
			assertTaskCount(t, tm, initialCount+1)

			// Vérifier la dernière tâche ajoutée
			lastTask := tm.Tasks[len(tm.Tasks)-1]
			if lastTask.Text != tt.text {
				t.Errorf("Texte attendu: %s, obtenu: %s", tt.text, lastTask.Text)
			}
			if len(lastTask.Tags) != len(tt.tags) {
				t.Errorf("Nombre de tags attendu: %d, obtenu: %d", len(tt.tags), len(lastTask.Tags))
			}
			if lastTask.UUID == "" {
				t.Error("UUID ne doit pas être vide")
			}
			if lastTask.Done {
				t.Error("Nouvelle tâche ne doit pas être marquée comme terminée")
			}
		})
	}
}

func TestTodoManager_Done(t *testing.T) {
	t.Run("marquer tâche existante comme terminée", func(t *testing.T) {
		tm, _, cleanup := setupTestEnvironment(t)
		defer cleanup()

		// Ajouter une tâche de test
		tm.Add("Test task", []string{}, "", "")
		taskID := tm.Tasks[0].ID

		// Marquer comme terminée
		tm.Done(taskID)

		// Vérifier que la tâche est marquée comme terminée
		task := assertTaskExists(t, tm, taskID)
		if task != nil && !task.Done {
			t.Error("Tâche devrait être marquée comme terminée")
		}
	})

	t.Run("marquer tâche inexistante", func(t *testing.T) {
		tm, _, cleanup := setupTestEnvironment(t)
		defer cleanup()

		// Essayer de marquer une tâche qui n'existe pas
		tm.Done(999)
		// Pas d'erreur attendue, juste un message affiché
	})
}

func TestTodoManager_Remove(t *testing.T) {
	t.Run("supprimer tâche existante", func(t *testing.T) {
		tm, _, cleanup := setupTestEnvironment(t)
		defer cleanup()

		// Ajouter deux tâches
		tm.Add("Task 1", []string{}, "", "")
		tm.Add("Task 2", []string{}, "", "")
		initialCount := len(tm.Tasks)

		// Supprimer la première tâche
		taskID := tm.Tasks[0].ID
		tm.Remove(taskID)

		// Vérifier que le nombre a diminué
		assertTaskCount(t, tm, initialCount-1)

		// Vérifier que la tâche n'existe plus
		for _, task := range tm.Tasks {
			if task.ID == taskID {
				t.Error("Tâche supprimée ne devrait plus exister")
			}
		}
	})
}

func TestTodoManager_Edit(t *testing.T) {
	t.Run("modifier tâche existante", func(t *testing.T) {
		tm, _, cleanup := setupTestEnvironment(t)
		defer cleanup()

		// Ajouter une tâche
		tm.Add("Original text", []string{"+old"}, "", "")
		taskID := tm.Tasks[0].ID
		originalUpdated := tm.Tasks[0].Updated

		// Attendre pour garantir une différence de timestamp
		//time.Sleep(time.Millisecond * 10)
		time.Sleep(2 * time.Second)

		// Modifier la tâche
		newText := "Modified text"
		newTags := []string{"+new", "@context"}
		tm.Edit(taskID, newText, newTags)

		// Vérifier les modifications
		task := assertTaskExists(t, tm, taskID)
		if task != nil {
			if task.Text != newText {
				t.Errorf("Texte attendu: %s, obtenu: %s", newText, task.Text)
			}
			if len(task.Tags) != len(newTags) {
				t.Errorf("Nombre de tags attendu: %d, obtenu: %d", len(newTags), len(task.Tags))
			}
			if task.Updated == originalUpdated {
				t.Error("Timestamp Updated devrait être mis à jour")
			}
		}
	})
}

func TestTodoManager_SaveLoad(t *testing.T) {
	t.Run("cycle sauvegarde et chargement", func(t *testing.T) {
		tm, _, cleanup := setupTestEnvironment(t)
		defer cleanup()

		// Ajouter des tâches
		tm.Add("Task 1", []string{"+test"}, "high", "2025-07-20")
		tm.Add("Task 2", []string{"@context"}, "medium", "")

		// Sauvegarder
		err := tm.save()
		if err != nil {
			t.Fatalf("Erreur lors de la sauvegarde: %v", err)
		}

		// Vérifier que le fichier existe
		if _, err := os.Stat(tm.filename); os.IsNotExist(err) {
			t.Fatal("Fichier de sauvegarde n'existe pas")
		}

		// Créer un nouveau TodoManager et charger
		tm2 := &TodoManager{
			Tasks:    []Task{},
			NextID:   1,
			filename: tm.filename,
		}
		tm2.load()

		// Vérifier que les données sont identiques
		if len(tm2.Tasks) != len(tm.Tasks) {
			t.Errorf("Nombre de tâches après chargement: attendu %d, obtenu %d",
				len(tm.Tasks), len(tm2.Tasks))
		}
		if tm2.NextID != tm.NextID {
			t.Errorf("NextID après chargement: attendu %d, obtenu %d",
				tm.NextID, tm2.NextID)
		}
	})

	t.Run("chargement fichier inexistant", func(t *testing.T) {
		tm := &TodoManager{
			Tasks:    []Task{},
			NextID:   1,
			filename: "/path/that/does/not/exist/todo.json",
		}

		// Le chargement ne doit pas échouer
		tm.load()

		// Les valeurs par défaut doivent être conservées
		assertTaskCount(t, tm, 0)
		if tm.NextID != 1 {
			t.Errorf("NextID devrait rester 1, obtenu: %d", tm.NextID)
		}
	})
}

func TestTodoManager_List(t *testing.T) {
	tm, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Ajouter des tâches de test avec différents états
	tm.Tasks = createSampleTasks()
	tm.NextID = 3

	t.Run("lister toutes les tâches actives", func(t *testing.T) {
		// Capturer la sortie (dans un vrai test, utiliser un buffer)
		tm.List(false, "", "", "")
		// Ici, on vérifierait que seules les tâches non terminées sont affichées
	})

	t.Run("lister avec filtre par projet", func(t *testing.T) {
		tm.List(false, "dev", "", "")
		// Vérifier que seules les tâches avec tag +dev sont considérées
	})

	t.Run("lister avec filtre par priorité", func(t *testing.T) {
		tm.List(false, "", "", "high")
		// Vérifier que seules les tâches avec priorité high sont considérées
	})
}

// Tests des fonctions utilitaires

func TestGenerateUUID(t *testing.T) {
	t.Run("génération UUID valide", func(t *testing.T) {
		uuid1 := generateUUID()
		uuid2 := generateUUID()

		// Vérifier le format UUID
		if len(uuid1) != 36 {
			t.Errorf("Longueur UUID attendue: 36, obtenue: %d", len(uuid1))
		}

		// Vérifier l'unicité
		if uuid1 == uuid2 {
			t.Error("Les UUIDs générés doivent être uniques")
		}

		// Vérifier le format (8-4-4-4-12)
		expectedPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
		matched, _ := regexp.MatchString(expectedPattern, uuid1)
		if !matched {
			t.Errorf("UUID ne correspond pas au format attendu: %s", uuid1)
		}
	})
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"high", "high"},
		{"h", "high"},
		{"haute", "high"},
		{"HIGH", "high"},
		{"medium", "medium"},
		{"m", "medium"},
		{"low", "low"},
		{"l", "low"},
		{"invalid", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parsePriority(tt.input)
			if result != tt.expected {
				t.Errorf("parsePriority(%s): attendu %s, obtenu %s",
					tt.input, tt.expected, result)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"2025-07-20", true},
		{"2025-01-01", true},
		{"2025-13-01", false}, // Mois invalide
		{"2025-07-32", false}, // Jour invalide
		{"2025/07/20", false}, // Format invalide
		{"", true},            // Vide = valide
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := validateDate(tt.input)
			if result != tt.expected {
				t.Errorf("validateDate(%s): attendu %t, obtenu %t",
					tt.input, tt.expected, result)
			}
		})
	}
}

func TestFilterTasks(t *testing.T) {
	tm, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Utiliser les tâches d'exemple
	tm.Tasks = createSampleTasks()

	t.Run("filtrer par projet", func(t *testing.T) {
		filtered := tm.filterTasks(false, "dev", "", "")
		expectedCount := 1 // Seule la première tâche a +dev
		if len(filtered) != expectedCount {
			t.Errorf("Filtre projet 'dev': attendu %d tâches, obtenu %d",
				expectedCount, len(filtered))
		}
	})

	t.Run("filtrer par contexte", func(t *testing.T) {
		filtered := tm.filterTasks(false, "", "bureau", "")
		expectedCount := 1 // Seule la première tâche a @bureau
		if len(filtered) != expectedCount {
			t.Errorf("Filtre contexte 'bureau': attendu %d tâches, obtenu %d",
				expectedCount, len(filtered))
		}
	})

	t.Run("filtrer par priorité", func(t *testing.T) {
		filtered := tm.filterTasks(false, "", "", "high")
		expectedCount := 1 // Seule la première tâche a priority=high
		if len(filtered) != expectedCount {
			t.Errorf("Filtre priorité 'high': attendu %d tâches, obtenu %d",
				expectedCount, len(filtered))
		}
	})

	t.Run("afficher toutes les tâches", func(t *testing.T) {
		filtered := tm.filterTasks(true, "", "", "")
		expectedCount := 2 // Toutes les tâches, y compris terminées
		if len(filtered) != expectedCount {
			t.Errorf("Toutes les tâches: attendu %d, obtenu %d",
				expectedCount, len(filtered))
		}
	})

	t.Run("masquer les tâches terminées", func(t *testing.T) {
		filtered := tm.filterTasks(false, "", "", "")
		expectedCount := 1 // Seule la première tâche (Done=false)
		if len(filtered) != expectedCount {
			t.Errorf("Tâches actives: attendu %d, obtenu %d",
				expectedCount, len(filtered))
		}
	})
}

// Benchmark pour les performances

func BenchmarkTodoManager_Add(b *testing.B) {
	tm, _, cleanup := setupTestEnvironment(&testing.T{})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.Add("Benchmark task", []string{"+bench"}, "medium", "")
	}
}

func BenchmarkTodoManager_List(b *testing.B) {
	tm, _, cleanup := setupTestEnvironment(&testing.T{})
	defer cleanup()

	// Pré-remplir avec 1000 tâches
	for i := 0; i < 1000; i++ {
		tm.Add("Task "+string(rune(i)), []string{"+bench"}, "medium", "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.List(false, "", "", "")
	}
}

func BenchmarkGenerateUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateUUID()
	}
}

// Test d'intégration complet

func TestIntegration_CompleteWorkflow(t *testing.T) {
	tm, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("workflow complet", func(t *testing.T) {
		// 1. Ajouter plusieurs tâches
		tm.Add("Tâche 1", []string{"+dev", "@bureau"}, "high", "2025-07-20")
		tm.Add("Tâche 2", []string{"+perso"}, "medium", "")
		tm.Add("Tâche 3", []string{"+dev", "@maison"}, "low", "2025-07-25")

		assertTaskCount(t, tm, 3)

		// 2. Marquer une tâche comme terminée
		tm.Done(1)
		task1 := assertTaskExists(t, tm, 1)
		if task1 != nil && !task1.Done {
			t.Error("Tâche 1 devrait être marquée comme terminée")
		}

		// 3. Modifier une tâche
		tm.Edit(2, "Tâche 2 modifiée", []string{"+perso", "@maison"})
		task2 := assertTaskExists(t, tm, 2)
		if task2 != nil && task2.Text != "Tâche 2 modifiée" {
			t.Error("Tâche 2 devrait être modifiée")
		}

		// 4. Export CSV
		csvFile := filepath.Join(tempDir, "export.csv")
		err := tm.ExportCSV(csvFile)
		if err != nil {
			t.Fatalf("Erreur lors de l'export: %v", err)
		}

		// Vérifier que le fichier CSV existe et contient les bonnes données
		content, err := ioutil.ReadFile(csvFile)
		if err != nil {
			t.Fatalf("Impossible de lire le fichier CSV: %v", err)
		}
		csvContent := string(content)
		if !strings.Contains(csvContent, "Tâche 1") {
			t.Error("Le CSV devrait contenir 'Tâche 1'")
		}

		// 5. Supprimer une tâche
		tm.Remove(3)
		assertTaskCount(t, tm, 2)

		// 6. Test de persistance
		err = tm.save()
		if err != nil {
			t.Fatalf("Erreur lors de la sauvegarde: %v", err)
		}

		// Créer un nouveau manager et charger
		tm2 := &TodoManager{
			Tasks:    []Task{},
			NextID:   1,
			filename: tm.filename,
		}
		tm2.load()

		if len(tm2.Tasks) != len(tm.Tasks) {
			t.Errorf("Nombre de tâches après rechargement: attendu %d, obtenu %d",
				len(tm.Tasks), len(tm2.Tasks))
		}
	})
}

// Tests d'import/export CSV

func TestTodoManager_ExportCSV(t *testing.T) {
	tm, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Ajouter des tâches de test
	tm.Tasks = createSampleTasks()

	csvFile := filepath.Join(tempDir, "test_export.csv")

	t.Run("export CSV réussi", func(t *testing.T) {
		err := tm.ExportCSV(csvFile)
		if err != nil {
			t.Fatalf("Erreur lors de l'export: %v", err)
		}

		// Vérifier que le fichier existe
		if _, err := os.Stat(csvFile); os.IsNotExist(err) {
			t.Fatal("Fichier CSV n'a pas été créé")
		}

		// Lire et vérifier le contenu
		content, err := ioutil.ReadFile(csvFile)
		if err != nil {
			t.Fatalf("Impossible de lire le fichier CSV: %v", err)
		}

		csvContent := string(content)

		// Vérifier la présence de l'en-tête
		if !strings.Contains(csvContent, "ID,UUID,Text,Done,Priority,Due,Tags,Created,Updated") {
			t.Error("En-tête CSV incorrect")
		}

		// Vérifier la présence des données
		if !strings.Contains(csvContent, "Tâche de test 1") {
			t.Error("Données de tâche manquantes dans le CSV")
		}

		// Compter les lignes (en-tête + 2 tâches = 3 lignes)
		lines := strings.Split(strings.TrimSpace(csvContent), "\n")
		expectedLines := 3
		if len(lines) != expectedLines {
			t.Errorf("Nombre de lignes CSV: attendu %d, obtenu %d", expectedLines, len(lines))
		}
	})

	t.Run("export vers fichier non accessible", func(t *testing.T) {
		invalidPath := "/root/non_accessible/test.csv"
		err := tm.ExportCSV(invalidPath)
		if err == nil {
			t.Error("Export devrait échouer avec un chemin non accessible")
		}
	})
}
func TestTodoManager_ImportCSV(t *testing.T) {
	tm, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Créer un fichier CSV avec des données valides
	csvContent := `ID,UUID,Text,Done,Priority,Due,Tags,Created,Updated
1,123e4567-e89b-42d3-a456-426614174000,Tâche importée 1,false,high,2025-07-25,"+import @test",2025-07-09 12:00:00,2025-07-09 12:00:00
2,123e4567-e89b-42d3-a456-426614174002,Tâche importée 2,true,medium,,@test,2025-07-09 13:00:00,2025-07-09 13:00:00`

	csvFile := filepath.Join(tempDir, "import_test.csv")
	err := ioutil.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Impossible de créer le fichier CSV de test: %v", err)
	}

	t.Run("import réussi en mode merge", func(t *testing.T) {
		options := ImportOptions{DryRun: false, Verbose: true}
		result, err := tm.ImportCSV(csvFile, "merge", "skip", options)

		if err != nil {
			t.Fatalf("Erreur lors de l'import: %v", err)
		}

		if result.NewTasks != 2 {
			t.Errorf("Nouvelles tâches attendues: 2, obtenues: %d", result.NewTasks)
		}

		assertTaskCount(t, tm, 2)

		// Vérifier les données importées
		task1 := tm.Tasks[0]
		if task1.Text != "Tâche importée 1" {
			t.Errorf("Texte de tâche incorrect: %s", task1.Text)
		}
		if task1.Priority != "high" {
			t.Errorf("Priorité incorrecte: %s", task1.Priority)
		}
	})

	t.Run("import avec conflit UUID - stratégie skip", func(t *testing.T) {
		// Préparer une tâche existante avec même UUID
		tm.Tasks = []Task{{
			ID:      1,
			UUID:    "123e4567-e89b-42d3-a456-426614174000",
			Text:    "Tâche existante",
			Done:    false,
			Created: "2025-07-09 10:00:00",
			Updated: "2025-07-09 10:00:00",
		}}
		tm.NextID = 2

		options := ImportOptions{DryRun: false, Verbose: false}
		result, err := tm.ImportCSV(csvFile, "merge", "skip", options)

		if err != nil {
			t.Fatalf("Erreur lors de l'import: %v", err)
		}

		if result.SkippedTasks != 1 {
			t.Errorf("Tâches ignorées attendues: 1, obtenues: %d", result.SkippedTasks)
		}
	})

	t.Run("import avec conflit UUID - stratégie update", func(t *testing.T) {
		// Réinitialiser l'état
		tm.Tasks = []Task{{
			ID:      1,
			UUID:    "123e4567-e89b-42d3-a456-426614174000",
			Text:    "Tâche existante",
			Done:    false,
			Created: "2025-07-09 10:00:00",
			Updated: "2025-07-09 10:00:00",
		}}
		tm.NextID = 2

		options := ImportOptions{DryRun: false, Verbose: false}
		result, err := tm.ImportCSV(csvFile, "merge", "update", options)

		if err != nil {
			t.Fatalf("Erreur lors de l'import: %v", err)
		}

		if result.UpdatedTasks != 1 {
			t.Errorf("Tâches mises à jour attendues: 1, obtenues: %d", result.UpdatedTasks)
		}
	})

	t.Run("import fichier inexistant", func(t *testing.T) {
		options := ImportOptions{DryRun: false, Verbose: false}
		_, err := tm.ImportCSV("fichier_inexistant.csv", "merge", "skip", options)

		if err == nil {
			t.Error("Import devrait échouer avec un fichier inexistant")
		}
	})

	t.Run("import dry-run", func(t *testing.T) {
		tm.Tasks = []Task{} // Réinitialiser
		tm.NextID = 1

		options := ImportOptions{DryRun: true, Verbose: false}
		result, err := tm.ImportCSV(csvFile, "merge", "skip", options)

		if err != nil {
			t.Fatalf("Erreur lors de l'import dry-run: %v", err)
		}

		// Rien ne doit être ajouté
		assertTaskCount(t, tm, 0)

		// Mais le résultat doit indiquer ce qui aurait été fait
		if result.NewTasks != 2 {
			t.Errorf("Dry-run devrait indiquer 2 nouvelles tâches, obtenu: %d", result.NewTasks)
		}
	})
}

/*func TestTodoManager_ImportCSV(t *testing.T) {
	tm, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// CORRECTION: Créer un fichier CSV avec un format plus simple et valide
	csvContent := `ID,UUID,Text,Done,Priority,Due,Tags,Created,Updated
1,123e4567-e89b-42d3-a456-426614174000,Tâche importée 1,false,high,2025-07-25,"+import @test",2025-07-09 12:00:00,2025-07-09 12:00:00
2,123e4567-e89b-42d3-a456-426614174002,Tâche importée 2,true,medium,,@test,2025-07-09 13:00:00,2025-07-09 13:00:00`


	csvFile := filepath.Join(tempDir, "import_test.csv")
	err := ioutil.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Impossible de créer le fichier CSV de test: %v", err)
	}

	t.Run("import réussi en mode merge", func(t *testing.T) {
		options := ImportOptions{DryRun: false, Verbose: true}
		result, err := tm.ImportCSV(csvFile, "merge", "skip", options)

		if err != nil {
			t.Fatalf("Erreur lors de l'import: %v", err)
		}

		if result.NewTasks != 2 {
			t.Errorf("Nouvelles tâches attendues: 2, obtenues: %d", result.NewTasks)
		}

		assertTaskCount(t, tm, 2)

		// Vérifier les données importées
		if len(tm.Tasks) > 0 {
			task1 := tm.Tasks[0]
			if task1.Text != "Tâche importée 1" {
				t.Errorf("Texte de tâche incorrect: %s", task1.Text)
			}
			if task1.Priority != "high" {
				t.Errorf("Priorité incorrecte: %s", task1.Priority)
			}
		}
	})

	t.Run("import avec conflit UUID", func(t *testing.T) {
		// Ajouter une tâche avec le même UUID
		tm.Tasks = []Task{{
			ID:      1,
			UUID:    "123e4567-e89b-42d3-a456-426614174000",
			Text:    "Tâche existante",
			Done:    false,
			Created: "2025-07-09 10:00:00",
			Updated: "2025-07-09 10:00:00",
		}}
		tm.NextID = 2

		options := ImportOptions{DryRun: false, Verbose: false}

		// Test stratégie skip
		result, err := tm.ImportCSV(csvFile, "merge", "skip", options)
		if err != nil {
			t.Logf("AA0")
			t.Fatalf("Erreur lors de l'import: %v", err)
		}

		if result.SkippedTasks != 1 {
			t.Logf("AA1")
			t.Errorf("Tâches ignorées attendues: 1, obtenues: %d", result.SkippedTasks)
		}

		tm.Tasks = []Task{{
			ID:      1,
			UUID:    "123e4567-e89b-42d3-a456-426614174000",
			Text:    "Tâche existante",
			Done:    false,
			Created: "2025-07-09 10:00:00",
			Updated: "2025-07-09 10:00:00",
		}}
		tm.NextID = 2

		// Test stratégie update
		result, err = tm.ImportCSV(csvFile, "merge", "update", options)
		if err != nil {
			t.Logf("AA2")
			t.Fatalf("Erreur lors de l'import: %v", err)
		}

		if result.UpdatedTasks != 1 {
			t.Logf("AA3")
			t.Errorf("Tâches mises à jour attendues: 1, obtenues: %d", result.UpdatedTasks)
		}
	})

	t.Run("import fichier inexistant", func(t *testing.T) {
		options := ImportOptions{DryRun: false, Verbose: false}
		_, err := tm.ImportCSV("fichier_inexistant.csv", "merge", "skip", options)

		if err == nil {
			t.Error("Import devrait échouer avec un fichier inexistant")
		}
	})

	t.Run("import dry-run", func(t *testing.T) {
		tm.Tasks = []Task{} // Réinitialiser
		tm.NextID = 1

		options := ImportOptions{DryRun: true, Verbose: false}
		result, err := tm.ImportCSV(csvFile, "merge", "skip", options)

		if err != nil {
			t.Fatalf("Erreur lors de l'import dry-run: %v", err)
		}

		// Aucune tâche ne devrait être ajoutée en mode dry-run
		assertTaskCount(t, tm, 0)

		// Mais le résultat devrait indiquer les tâches qui auraient été ajoutées
		if result.NewTasks != 2 {
			t.Errorf("Dry-run devrait indiquer 2 nouvelles tâches, obtenu: %d", result.NewTasks)
		}
	})
}*/

// Tests de validation et d'erreurs

func TestErrorHandling(t *testing.T) {
	t.Run("fichier JSON corrompu", func(t *testing.T) {
		tm, tempDir, cleanup := setupTestEnvironment(t)
		defer cleanup()
		_ = tempDir // Éviter l'erreur "variable non utilisée"

		// Créer un fichier JSON invalide
		corruptedJSON := `{"tasks": [invalid json}`
		err := ioutil.WriteFile(tm.filename, []byte(corruptedJSON), 0644)
		if err != nil {
			t.Fatalf("Impossible de créer le fichier JSON corrompu: %v", err)
		}

		// Le chargement ne devrait pas planter
		tm.load()

		// Les valeurs par défaut devraient être conservées
		assertTaskCount(t, tm, 0)
		if tm.NextID != 1 {
			t.Errorf("NextID devrait être 1 après chargement échec, obtenu: %d", tm.NextID)
		}
	})

	t.Run("validation des entrées", func(t *testing.T) {
		tm, _, cleanup := setupTestEnvironment(t)
		defer cleanup()

		// Test avec texte vide (devrait être autorisé)
		tm.Add("", []string{}, "", "")
		assertTaskCount(t, tm, 1)

		// Test avec date invalide (validée par validateDate)
		invalidDate := "2025-13-40"
		isValid := validateDate(invalidDate)
		if isValid {
			t.Error("Date invalide ne devrait pas être acceptée")
		}
	})
}

// Tests de performance

func TestPerformance_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Test de performance ignoré en mode court")
	}

	tm, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("performance avec 1000 tâches", func(t *testing.T) {
		// Ajouter 1000 tâches
		start := time.Now()
		for i := 0; i < 1000; i++ {
			tm.Add(fmt.Sprintf("Tâche %d", i), []string{"+perf"}, "medium", "")
		}
		addDuration := time.Since(start)

		t.Logf("Temps pour ajouter 1000 tâches: %v", addDuration)
		assertTaskCount(t, tm, 1000)

		// Test de listing
		start = time.Now()
		tm.List(false, "", "", "")
		listDuration := time.Since(start)
		t.Logf("Temps pour lister 1000 tâches: %v", listDuration)

		// Test de sauvegarde
		start = time.Now()
		err := tm.save()
		if err != nil {
			t.Fatalf("Erreur lors de la sauvegarde: %v", err)
		}
		saveDuration := time.Since(start)
		t.Logf("Temps pour sauvegarder 1000 tâches: %v", saveDuration)

		// CORRECTION: Limites de performance plus réalistes
		if addDuration > 5*time.Second {
			t.Errorf("Ajout de 1000 tâches trop lent: %v", addDuration)
		}
		if listDuration > 100*time.Millisecond {
			t.Errorf("Listing de 1000 tâches trop lent: %v", listDuration)
		}
		if saveDuration > time.Second {
			t.Errorf("Sauvegarde de 1000 tâches trop lente: %v", saveDuration)
		}
	})
}

// Tests spécifiques aux fonctions d'import

func TestImportHelperFunctions(t *testing.T) {
	tm, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("isValidUUID", func(t *testing.T) {
		tests := []struct {
			uuid     string
			expected bool
		}{
			{"123e4567-e89b-12d3-a456-426614174000", false}, // Ce n'est pas un UUID v4 valide
			{"123e4567-e89b-42d3-a456-426614174000", true},  // UUID v4 valide
			{"invalid-uuid", false},
			{"", false},
			{"550e8400-e29b-41d4-a716-446655440000", true},  // UUID v4 valide
			{"123e4567-e89b-12d3-g456-426614174000", false}, // Caractère invalide
		}

		for _, tt := range tests {
			result := tm.isValidUUID(tt.uuid)
			if result != tt.expected {
				t.Errorf("isValidUUID(%s): attendu %t, obtenu %t",
					tt.uuid, tt.expected, result)
			}
		}
	})

	t.Run("isValidDateTime", func(t *testing.T) {
		tests := []struct {
			datetime string
			expected bool
		}{
			{"2025-07-09 15:04:05", true},
			{"2025-07-09T15:04:05", true},
			{"2025-07-09 15:04", true},
			{"2025-07-09", false}, // Pas de format supporté
			{"invalid", false},
			{"", false},
		}

		for _, tt := range tests {
			result := tm.isValidDateTime(tt.datetime)
			if result != tt.expected {
				t.Errorf("isValidDateTime(%s): attendu %t, obtenu %t",
					tt.datetime, tt.expected, result)
			}
		}
	})

	t.Run("parseTags", func(t *testing.T) {
		tests := []struct {
			input    string
			expected []string
		}{
			{"+dev @bureau", []string{"+dev", "@bureau"}},
			{"+projet1 +projet2 @context", []string{"+projet1", "+projet2", "@context"}},
			{"invalid tag +valid @valid", []string{"+valid", "@valid"}},
			{"", []string{}},
			{"no tags here", []string{}},
		}

		for _, tt := range tests {
			result := tm.parseTags(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseTags(%s): nombre attendu %d, obtenu %d",
					tt.input, len(tt.expected), len(result))
				continue
			}
			for i, tag := range tt.expected {
				if i >= len(result) || result[i] != tag {
					t.Errorf("parseTags(%s): tag %d attendu %s, obtenu %s",
						tt.input, i, tag, result[i])
				}
			}
		}
	})

	t.Run("isNewer", func(t *testing.T) {
		tests := []struct {
			date1    string
			date2    string
			expected bool
		}{
			{"2025-07-09 15:00:00", "2025-07-09 14:00:00", true},
			{"2025-07-09 14:00:00", "2025-07-09 15:00:00", false},
			{"2025-07-09 15:00:00", "2025-07-09 15:00:00", false},
			{"invalid", "2025-07-09 15:00:00", false},
			{"2025-07-09 15:00:00", "invalid", false},
		}

		for _, tt := range tests {
			result := tm.isNewer(tt.date1, tt.date2)
			if result != tt.expected {
				t.Errorf("isNewer(%s, %s): attendu %t, obtenu %t",
					tt.date1, tt.date2, tt.expected, result)
			}
		}
	})
}

// Test de round-trip export/import
func TestExportImportRoundTrip(t *testing.T) {
	tm1, tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Ajouter des tâches variées dans le premier manager
	tm1.Add("Tâche avec émojis 🚀", []string{"+emoji", "@test"}, "high", "2025-07-20")
	tm1.Add("Tâche avec \"guillemets\"", []string{"+quotes"}, "medium", "")
	tm1.Add("Tâche avec caractères spéciaux àéèç", []string{"+unicode"}, "low", "2025-12-31")
	tm1.Done(2) // Marquer la deuxième comme terminée

	// Export
	csvFile := filepath.Join(tempDir, "roundtrip.csv")
	err := tm1.ExportCSV(csvFile)
	if err != nil {
		t.Fatalf("Erreur lors de l'export: %v", err)
	}

	// Créer un nouveau manager et importer
	tm2, _, cleanup2 := setupTestEnvironment(t)
	defer cleanup2()

	options := ImportOptions{DryRun: false, Verbose: false}
	result, err := tm2.ImportCSV(csvFile, "merge", "skip", options)
	if err != nil {
		t.Fatalf("Erreur lors de l'import: %v", err)
	}

	// Vérifications
	if result.NewTasks != len(tm1.Tasks) {
		t.Errorf("Nombre de tâches importées: attendu %d, obtenu %d",
			len(tm1.Tasks), result.NewTasks)
	}

	assertTaskCount(t, tm2, len(tm1.Tasks))

	// Vérifier que les données sont identiques
	for i, originalTask := range tm1.Tasks {
		importedTask := tm2.Tasks[i]

		if importedTask.Text != originalTask.Text {
			t.Errorf("Texte différent pour tâche %d: %s vs %s",
				i, originalTask.Text, importedTask.Text)
		}
		if importedTask.Done != originalTask.Done {
			t.Errorf("Statut Done différent pour tâche %d", i)
		}
		if importedTask.Priority != originalTask.Priority {
			t.Errorf("Priorité différente pour tâche %d", i)
		}
		if len(importedTask.Tags) != len(originalTask.Tags) {
			t.Errorf("Nombre de tags différent pour tâche %d", i)
		}
	}
}
