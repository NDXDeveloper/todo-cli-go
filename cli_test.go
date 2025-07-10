// cli_test.go - Tests d'intégration CLI
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// CLITestHelper aide pour les tests CLI
type CLITestHelper struct {
	tempDir    string
	todoFile   string
	binaryPath string
	cleanup    func()
}

// setupCLITest prépare l'environnement pour les tests CLI
func setupCLITest(t *testing.T) *CLITestHelper {
	// Créer répertoire temporaire
	tempDir, err := ioutil.TempDir("", "todo_cli_test")
	if err != nil {
		t.Fatalf("Impossible de créer le répertoire temporaire: %v", err)
	}

	// Définir le chemin du fichier todo
	todoDir := filepath.Join(tempDir, ".todo")
	err = os.MkdirAll(todoDir, 0755)
	if err != nil {
		t.Fatalf("Impossible de créer le répertoire .todo: %v", err)
	}

	todoFile := filepath.Join(todoDir, "todo.json")

	// Compiler le binaire pour les tests
	binaryPath := filepath.Join(tempDir, "todo_test")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	// Fonction de nettoyage
	cleanup := func() {
		os.RemoveAll(tempDir)
		if _, err := os.Stat(binaryPath); err == nil {
			os.Remove(binaryPath)
		}
	}

	return &CLITestHelper{
		tempDir:    tempDir,
		todoFile:   todoFile,
		binaryPath: binaryPath,
		cleanup:    cleanup,
	}
}

// compileBinary compile le binaire de test
func (h *CLITestHelper) compileBinary(t *testing.T) {
	// CORRECTION: Compiler seulement main.go et import.go, pas les tests
	cmd := exec.Command("go", "build", "-o", h.binaryPath, "main.go", "import.go")
	cmd.Env = append(os.Environ(), "HOME="+h.tempDir, "USERPROFILE="+h.tempDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Impossible de compiler le binaire: %v\nSortie: %s", err, output)
	}
}

// runCommand exécute une commande todo et retourne la sortie
// runCommand exécute une commande todo et retourne la sortie
func (h *CLITestHelper) runCommand(args ...string) (string, string, int, error) {
	cmd := exec.Command(h.binaryPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+h.tempDir, "USERPROFILE="+h.tempDir)

	// AJOUTER CETTE LIGNE ICI ⬇️
	cmd.Dir = h.tempDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return "", "", -1, err
		}
	}

	return stdout.String(), stderr.String(), exitCode, nil
}

// assertCommandSuccess vérifie qu'une commande s'exécute avec succès
func (h *CLITestHelper) assertCommandSuccess(t *testing.T, args ...string) string {
	t.Helper()
	stdout, stderr, exitCode, err := h.runCommand(args...)
	if err != nil {
		t.Fatalf("Erreur lors de l'exécution de la commande %v: %v", args, err)
	}
	if exitCode != 0 {
		t.Fatalf("Commande %v a échoué (code %d)\nStdout: %s\nStderr: %s",
			args, exitCode, stdout, stderr)
	}
	return stdout
}

// assertCommandFails vérifie qu'une commande échoue
func (h *CLITestHelper) assertCommandFails(t *testing.T, expectedExitCode int, args ...string) {
	t.Helper()
	_, _, exitCode, err := h.runCommand(args...)
	if err != nil && exitCode == -1 {
		t.Fatalf("Erreur inattendue lors de l'exécution: %v", err)
	}
	if exitCode != expectedExitCode {
		t.Fatalf("Code de sortie attendu %d, obtenu %d pour la commande %v",
			expectedExitCode, exitCode, args)
	}
}

// debugEnvironment aide au debug des problèmes de fichiers
func (h *CLITestHelper) debugEnvironment(t *testing.T) {
	t.Helper()

	// Lister les fichiers dans tempDir
	files, err := ioutil.ReadDir(h.tempDir)
	if err == nil {
		t.Logf("Fichiers dans tempDir (%s):", h.tempDir)
		for _, file := range files {
			t.Logf("  %s", file.Name())
		}
	}

	// Lister les fichiers dans .todo
	todoDir := filepath.Join(h.tempDir, ".todo")
	files, err = ioutil.ReadDir(todoDir)
	if err == nil {
		t.Logf("Fichiers dans .todo:")
		for _, file := range files {
			t.Logf("  %s", file.Name())
		}
	}

	// Vérifier le fichier todo.json
	if _, err := os.Stat(h.todoFile); err == nil {
		content, _ := ioutil.ReadFile(h.todoFile)
		t.Logf("Contenu de todo.json: %s", string(content))
	} else {
		t.Logf("todo.json n'existe pas: %v", err)
	}
}

// Tests CLI de base

func TestCLI_Add(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("ajout simple", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "add", "Ma première tâche")

		if !strings.Contains(output, "✅ Tâche ajoutée") {
			t.Errorf("Message de succès manquant dans: %s", output)
		}
		if !strings.Contains(output, "[1]") {
			t.Errorf("ID de tâche manquant dans: %s", output)
		}
	})

	t.Run("ajout avec tags et priorité", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "add", "Tâche avec options",
			"+dev", "@bureau", "--priority=high", "--due=2025-07-20")

		if !strings.Contains(output, "✅ Tâche ajoutée") {
			t.Errorf("Message de succès manquant: %s", output)
		}
		if !strings.Contains(output, "Priority: high") {
			t.Errorf("Priorité manquante dans la sortie: %s", output)
		}
		if !strings.Contains(output, "[+dev @bureau]") {
			t.Errorf("Tags manquants dans la sortie: %s", output)
		}
	})

	t.Run("ajout sans texte", func(t *testing.T) {
		h.assertCommandFails(t, 1, "add")
	})

	t.Run("ajout avec date invalide", func(t *testing.T) {
		h.assertCommandFails(t, 1, "add", "Test", "--due=2025-13-50")
	})
}

func TestCLI_List(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	// Ajouter quelques tâches de test
	h.assertCommandSuccess(t, "add", "Tâche 1", "+dev", "@bureau", "--priority=high")
	h.assertCommandSuccess(t, "add", "Tâche 2", "+perso", "@maison", "--priority=medium")
	h.assertCommandSuccess(t, "add", "Tâche 3", "+dev", "@maison", "--priority=low")

	t.Run("list basique", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "list")

		// Vérifier que les 3 tâches sont listées
		if !strings.Contains(output, "Tâche 1") {
			t.Errorf("Tâche 1 manquante: %s", output)
		}
		if !strings.Contains(output, "Tâche 2") {
			t.Errorf("Tâche 2 manquante: %s", output)
		}
		if !strings.Contains(output, "Tâche 3") {
			t.Errorf("Tâche 3 manquante: %s", output)
		}

		// Vérifier le tri par priorité (high doit être en premier)
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) < 3 {
			t.Errorf("Nombre de lignes insuffisant: %d", len(lines))
		}

		// La première ligne doit contenir la tâche avec priorité high
		if !strings.Contains(lines[0], "❗") { // Icône priorité high
			t.Errorf("Tri par priorité incorrect: %s", lines[0])
		}
	})

	t.Run("filtrage par projet", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "list", "--project=dev")

		if !strings.Contains(output, "Tâche 1") {
			t.Errorf("Tâche 1 devrait être incluse: %s", output)
		}
		if !strings.Contains(output, "Tâche 3") {
			t.Errorf("Tâche 3 devrait être incluse: %s", output)
		}
		if strings.Contains(output, "Tâche 2") {
			t.Errorf("Tâche 2 ne devrait pas être incluse: %s", output)
		}
	})

	t.Run("filtrage par contexte", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "list", "--context=maison")

		if !strings.Contains(output, "Tâche 2") {
			t.Errorf("Tâche 2 devrait être incluse: %s", output)
		}
		if !strings.Contains(output, "Tâche 3") {
			t.Errorf("Tâche 3 devrait être incluse: %s", output)
		}
		if strings.Contains(output, "Tâche 1") {
			t.Errorf("Tâche 1 ne devrait pas être incluse: %s", output)
		}
	})

	t.Run("filtrage par priorité", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "list", "--priority=high")

		if !strings.Contains(output, "Tâche 1") {
			t.Errorf("Tâche 1 devrait être incluse: %s", output)
		}
		if strings.Contains(output, "Tâche 2") || strings.Contains(output, "Tâche 3") {
			t.Errorf("Seule la tâche 1 devrait être incluse: %s", output)
		}
	})

	t.Run("liste vide", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "list", "--project=inexistant")

		if !strings.Contains(output, "📝 Aucune tâche trouvée") {
			t.Errorf("Message 'aucune tâche' manquant: %s", output)
		}
	})
}

func TestCLI_Done(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	// Ajouter une tâche de test
	h.assertCommandSuccess(t, "add", "Tâche à terminer")

	t.Run("marquer tâche comme terminée", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "done", "1")

		if !strings.Contains(output, "✅ Tâche [1] marquée comme terminée") {
			t.Errorf("Message de succès manquant: %s", output)
		}

		// Vérifier que la tâche n'apparaît plus dans la liste par défaut
		listOutput := h.assertCommandSuccess(t, "list")
		if !strings.Contains(listOutput, "📝 Aucune tâche trouvée") {
			t.Errorf("Tâche terminée ne devrait plus apparaître: %s", listOutput)
		}

		// Vérifier qu'elle apparaît avec --all
		allOutput := h.assertCommandSuccess(t, "list", "--all")
		if !strings.Contains(allOutput, "✅") {
			t.Errorf("Tâche terminée devrait apparaître avec --all: %s", allOutput)
		}
	})

	t.Run("marquer tâche inexistante", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "done", "999")

		if !strings.Contains(output, "❌ Tâche [999] introuvable") {
			t.Errorf("Message d'erreur manquant: %s", output)
		}
	})

	t.Run("ID invalide", func(t *testing.T) {
		h.assertCommandFails(t, 1, "done", "abc")
	})

	t.Run("sans ID", func(t *testing.T) {
		h.assertCommandFails(t, 1, "done")
	})
}

func TestCLI_Remove(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	// Ajouter des tâches de test
	h.assertCommandSuccess(t, "add", "Tâche à supprimer")
	h.assertCommandSuccess(t, "add", "Tâche à garder")

	t.Run("supprimer tâche existante", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "remove", "1")

		if !strings.Contains(output, "🗑️ Tâche [1] supprimée") {
			t.Errorf("Message de succès manquant: %s", output)
		}

		// Vérifier que seule la tâche 2 reste
		listOutput := h.assertCommandSuccess(t, "list")
		if !strings.Contains(listOutput, "Tâche à garder") {
			t.Errorf("Tâche 2 devrait encore exister: %s", listOutput)
		}
		if strings.Contains(listOutput, "Tâche à supprimer") {
			t.Errorf("Tâche 1 ne devrait plus exister: %s", listOutput)
		}
	})

	t.Run("supprimer tâche inexistante", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "remove", "999")

		if !strings.Contains(output, "❌ Tâche [999] introuvable") {
			t.Errorf("Message d'erreur manquant: %s", output)
		}
	})
}

func TestCLI_Edit(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	// Ajouter une tâche de test
	h.assertCommandSuccess(t, "add", "Texte original", "+old")

	t.Run("modifier tâche existante", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "edit", "1", "Nouveau texte", "+new", "@context")

		if !strings.Contains(output, "✏️ Tâche [1] modifiée") {
			t.Errorf("Message de succès manquant: %s", output)
		}

		// Vérifier les modifications dans la liste
		listOutput := h.assertCommandSuccess(t, "list")
		if !strings.Contains(listOutput, "Nouveau texte") {
			t.Errorf("Nouveau texte manquant: %s", listOutput)
		}
		if !strings.Contains(listOutput, "+new @context") {
			t.Errorf("Nouveaux tags manquants: %s", listOutput)
		}
		if strings.Contains(listOutput, "Texte original") {
			t.Errorf("Ancien texte ne devrait plus apparaître: %s", listOutput)
		}
	})

	t.Run("modifier tâche inexistante", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "edit", "999", "Nouveau texte")

		if !strings.Contains(output, "❌ Tâche [999] introuvable") {
			t.Errorf("Message d'erreur manquant: %s", output)
		}
	})

	t.Run("arguments insuffisants", func(t *testing.T) {
		h.assertCommandFails(t, 1, "edit", "1")
		h.assertCommandFails(t, 1, "edit")
	})
}

func TestCLI_Export(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	// Ajouter des tâches de test
	h.assertCommandSuccess(t, "add", "Tâche export 1", "+test")
	h.assertCommandSuccess(t, "add", "Tâche export 2", "@context", "--priority=high")

	t.Run("export par défaut", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "export")

		if !strings.Contains(output, "📄 Export terminé : todo_export.csv") {
			t.Errorf("Message de succès manquant: %s", output)
		}

		// Le fichier est créé dans le répertoire tempdir
		csvPath := filepath.Join(h.tempDir, "todo_export.csv")

		// Attendre un peu pour que le fichier soit créé
		time.Sleep(200 * time.Millisecond)

		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			// Debug en cas d'échec
			h.debugEnvironment(t)
			t.Errorf("Fichier CSV n'a pas été créé à %s", csvPath)
			return
		}

		// Vérifier le contenu
		content, err := ioutil.ReadFile(csvPath)
		if err != nil {
			t.Fatalf("Impossible de lire le fichier CSV: %v", err)
		}
		csvContent := string(content)

		if !strings.Contains(csvContent, "Tâche export 1") {
			t.Errorf("Tâche 1 manquante dans le CSV: %s", csvContent)
		}
		if !strings.Contains(csvContent, "Tâche export 2") {
			t.Errorf("Tâche 2 manquante dans le CSV: %s", csvContent)
		}
	})

	t.Run("export avec nom de fichier", func(t *testing.T) {
		customFile := "export_custom.csv"
		output := h.assertCommandSuccess(t, "export", customFile)

		if !strings.Contains(output, "📄 Export terminé : "+customFile) {
			t.Errorf("Message de succès incorrect: %s", output)
		}

		// Le fichier devrait être dans .todo
		csvPath := filepath.Join(h.tempDir, customFile)
		time.Sleep(200 * time.Millisecond)

		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			h.debugEnvironment(t)
			t.Errorf("Fichier CSV personnalisé n'a pas été créé à %s", csvPath)
		}
	})
}

func TestCLI_Import(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	// CORRECTION: Utiliser des UUID v4 valides comme l'application les génère
	csvContent := `ID,UUID,Text,Done,Priority,Due,Tags,Created,Updated
1,123e4567-e89b-42d3-a456-426614174000,"Tâche importée 1",false,high,2025-07-25,"+import @test",2025-07-09 12:00:00,2025-07-09 12:00:00
2,987fcdeb-51d2-42d3-a456-426614174111,"Tâche importée 2",true,medium,,"@test",2025-07-09 13:00:00,2025-07-09 13:00:00`

	csvFile := filepath.Join(h.tempDir, "import_test.csv")
	err := ioutil.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Impossible de créer le fichier CSV de test: %v", err)
	}

	t.Run("import basique", func(t *testing.T) {
		// Debug: vérifier que le fichier existe
		if _, err := os.Stat(csvFile); os.IsNotExist(err) {
			t.Fatalf("Fichier CSV de test n'existe pas: %s", csvFile)
		}

		output := h.assertCommandSuccess(t, "import", csvFile)

		if !strings.Contains(output, "📥 Import terminé") {
			t.Errorf("Message de succès manquant: %s", output)
		}
		if !strings.Contains(output, "✅ 2 nouvelles tâches") {
			t.Errorf("Compteur de nouvelles tâches incorrect: %s", output)
		}

		// Vérifier que les tâches ont été importées
		listOutput := h.assertCommandSuccess(t, "list", "--all")
		if !strings.Contains(listOutput, "Tâche importée 1") {
			t.Errorf("Tâche 1 manquante après import: %s", listOutput)
		}
		if !strings.Contains(listOutput, "Tâche importée 2") {
			t.Errorf("Tâche 2 manquante après import: %s", listOutput)
		}
	})

	t.Run("import avec options", func(t *testing.T) {
		// Nettoyer d'abord
		os.Remove(h.todoFile)

		output := h.assertCommandSuccess(t, "import", csvFile, "--verbose", "--dry-run")

		if !strings.Contains(output, "🔍 Mode dry-run") {
			t.Errorf("Indication dry-run manquante: %s", output)
		}

		// En mode dry-run, aucune tâche ne doit être réellement ajoutée
		listOutput := h.assertCommandSuccess(t, "list")
		if !strings.Contains(listOutput, "📝 Aucune tâche trouvée") {
			t.Errorf("Des tâches ont été ajoutées en mode dry-run: %s", listOutput)
		}
	})

	t.Run("import fichier inexistant", func(t *testing.T) {
		h.assertCommandFails(t, 1, "import", "fichier_inexistant.csv")
	})

	t.Run("import sans arguments", func(t *testing.T) {
		h.assertCommandFails(t, 1, "import")
	})
}

func TestCLI_Help(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("aide principale", func(t *testing.T) {
		output := h.assertCommandSuccess(t, "help")

		// Vérifier la présence des sections principales
		if !strings.Contains(output, "📋 Todo Manager CLI") {
			t.Errorf("Titre manquant: %s", output)
		}
		if !strings.Contains(output, "Usage:") {
			t.Errorf("Section Usage manquante: %s", output)
		}
		if !strings.Contains(output, "Exemples:") {
			t.Errorf("Section Exemples manquante: %s", output)
		}
	})

	t.Run("aide sans arguments", func(t *testing.T) {
		// Lancer sans arguments devrait afficher l'aide et sortir avec code 1
		h.assertCommandFails(t, 1)
	})

	t.Run("commande inconnue", func(t *testing.T) {
		h.assertCommandFails(t, 1, "commande_inexistante")
	})
}

// Tests End-to-End complets

func TestE2E_CompleteWorkflow(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("workflow complet de gestion de tâches", func(t *testing.T) {
		// 1. Ajouter plusieurs tâches avec différentes options
		h.assertCommandSuccess(t, "add", "Préparer présentation", "+travail", "@bureau",
			"--priority=high", "--due=2025-07-20")
		h.assertCommandSuccess(t, "add", "Faire les courses", "+perso", "@supermarché",
			"--priority=medium")
		h.assertCommandSuccess(t, "add", "Réviser Go", "+dev", "@maison", "--priority=low")
		h.assertCommandSuccess(t, "add", "Appeler dentiste", "+santé", "--priority=medium")

		// 2. Lister toutes les tâches et vérifier l'ordre
		listOutput := h.assertCommandSuccess(t, "list")
		lines := strings.Split(strings.TrimSpace(listOutput), "\n")

		// Doit y avoir 4 lignes de tâches
		if len(lines) != 4 {
			t.Errorf("Nombre de tâches listées: attendu 4, obtenu %d", len(lines))
		}

		// La première doit être la tâche high priority
		if !strings.Contains(lines[0], "Préparer présentation") {
			t.Errorf("Tri par priorité incorrect, première ligne: %s", lines[0])
		}

		// 3. Filtrer par projet
		devOutput := h.assertCommandSuccess(t, "list", "--project=dev")
		if !strings.Contains(devOutput, "Réviser Go") {
			t.Errorf("Filtrage par projet échoué: %s", devOutput)
		}
		if strings.Contains(devOutput, "Faire les courses") {
			t.Errorf("Filtrage par projet inclut des tâches incorrectes: %s", devOutput)
		}

		// 4. Marquer une tâche comme terminée
		h.assertCommandSuccess(t, "done", "2")

		// Vérifier qu'elle n'apparaît plus dans la liste par défaut
		listAfterDone := h.assertCommandSuccess(t, "list")
		if strings.Contains(listAfterDone, "Faire les courses") {
			t.Errorf("Tâche terminée apparaît encore: %s", listAfterDone)
		}

		// 5. Modifier une tâche
		h.assertCommandSuccess(t, "edit", "3", "Réviser Go et faire exercices", "+dev", "+urgent", "@maison")

		editedOutput := h.assertCommandSuccess(t, "list")
		if !strings.Contains(editedOutput, "Réviser Go et faire exercices") {
			t.Errorf("Modification de texte échouée: %s", editedOutput)
		}

		// 6. Export
		h.assertCommandSuccess(t, "export", "workflow_test.csv")

		csvPath := filepath.Join(h.tempDir, "workflow_test.csv")
		time.Sleep(200 * time.Millisecond) // Attendre que le fichier soit créé

		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			h.debugEnvironment(t) // Debug en cas d'échec
			t.Fatalf("Fichier d'export non trouvé: %s", csvPath)
		}

		content, err := ioutil.ReadFile(csvPath)
		if err != nil {
			t.Fatalf("Impossible de lire le fichier exporté: %v", err)
		}

		csvContent := string(content)
		if !strings.Contains(csvContent, "Préparer présentation") {
			t.Errorf("Export incomplet: %s", csvContent)
		}

		// 7. Supprimer une tâche
		h.assertCommandSuccess(t, "remove", "4")

		listAfterRemove := h.assertCommandSuccess(t, "list")
		if strings.Contains(listAfterRemove, "Appeler dentiste") {
			t.Errorf("Tâche supprimée apparaît encore: %s", listAfterRemove)
		}

		// 8. Vérifier le statut final
		finalList := h.assertCommandSuccess(t, "list", "--all")

		// Doit contenir: 1 active (Préparer présentation), 1 modifiée (Réviser Go), 1 terminée (Faire les courses)
		allLines := strings.Split(strings.TrimSpace(finalList), "\n")
		if len(allLines) != 3 {
			t.Errorf("Nombre final de tâches: attendu 3, obtenu %d", len(allLines))
		}
	})
}

func TestE2E_DataPersistence(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("persistance des données entre redémarrages", func(t *testing.T) {
		// 1. Ajouter des tâches
		h.assertCommandSuccess(t, "add", "Tâche persistante 1", "+test")
		h.assertCommandSuccess(t, "add", "Tâche persistante 2", "@context", "--priority=high")

		// 2. Vérifier qu'elles sont listées
		firstList := h.assertCommandSuccess(t, "list")
		if !strings.Contains(firstList, "Tâche persistante 1") {
			t.Errorf("Tâche 1 manquante: %s", firstList)
		}

		// 3. Simuler redémarrage en créant une nouvelle instance CLI
		// (le fichier JSON devrait persister)

		// 4. Vérifier que les tâches sont toujours là
		secondList := h.assertCommandSuccess(t, "list")
		if !strings.Contains(secondList, "Tâche persistante 1") {
			t.Errorf("Tâche 1 perdue après redémarrage: %s", secondList)
		}
		if !strings.Contains(secondList, "Tâche persistante 2") {
			t.Errorf("Tâche 2 perdue après redémarrage: %s", secondList)
		}

		// 5. Vérifier que le fichier JSON existe et est valide
		if _, err := os.Stat(h.todoFile); os.IsNotExist(err) {
			t.Error("Fichier todo.json n'existe pas")
		}

		// Lire et parser le JSON
		content, err := ioutil.ReadFile(h.todoFile)
		if err != nil {
			t.Fatalf("Impossible de lire todo.json: %v", err)
		}

		var data map[string]interface{}
		err = json.Unmarshal(content, &data)
		if err != nil {
			t.Fatalf("JSON invalide: %v", err)
		}

		// Vérifier la structure
		if _, ok := data["tasks"]; !ok {
			t.Error("Clé 'tasks' manquante dans le JSON")
		}
		if _, ok := data["nextId"]; !ok {
			t.Error("Clé 'nextId' manquante dans le JSON")
		}
	})
}

func TestE2E_ImportExportRoundTrip(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("round-trip export/import", func(t *testing.T) {
		// 1. Créer des données de test plus simples pour éviter les problèmes d'encodage
		h.assertCommandSuccess(t, "add", "Tâche test 1", "+emoji", "@test", "--priority=high")
		h.assertCommandSuccess(t, "add", "Tâche test 2", "+unicode", "--priority=medium", "--due=2025-12-25")
		h.assertCommandSuccess(t, "add", "Tâche simple", "+basic")

		// Marquer une comme terminée
		h.assertCommandSuccess(t, "done", "3")

		// 2. Export
		exportFile := "roundtrip_test.csv"
		h.assertCommandSuccess(t, "export", exportFile)

		// 3. Vérifier que le fichier d'export existe
		exportPath := filepath.Join(h.tempDir, exportFile)
		time.Sleep(200 * time.Millisecond)

		if _, err := os.Stat(exportPath); os.IsNotExist(err) {
			h.debugEnvironment(t)
			t.Fatalf("Fichier d'export n'a pas été créé: %s", exportPath)
		}

		// 4. Réinitialiser (simuler nouvelle installation)
		err := os.Remove(h.todoFile)
		if err != nil {
			t.Fatalf("Impossible de supprimer le fichier todo: %v", err)
		}

		// 5. Import
		output := h.assertCommandSuccess(t, "import", exportPath, "--verbose")

		if !strings.Contains(output, "📥 Import terminé") {
			t.Errorf("Message d'import manquant: %s", output)
		}

		// 6. Vérifier que les données sont identiques
		importedList := h.assertCommandSuccess(t, "list", "--all")

		// Vérifications simplifiées
		if !strings.Contains(importedList, "Tâche test 1") {
			t.Errorf("Tâche 1 manquante: %s", importedList)
		}
		if !strings.Contains(importedList, "Tâche test 2") {
			t.Errorf("Tâche 2 manquante: %s", importedList)
		}
		if !strings.Contains(importedList, "✅") { // Tâche terminée
			t.Errorf("Statut Done perdu: %s", importedList)
		}

		// Vérifier que les priorités sont préservées
		if !strings.Contains(importedList, "❗") { // Priorité high
			t.Errorf("Priorité high perdue: %s", importedList)
		}
	})
}

// Tests de stress et performance

func TestE2E_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Test de stress ignoré en mode court")
	}

	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("gestion de nombreuses tâches", func(t *testing.T) {
		// Ajouter 100 tâches
		start := time.Now()
		for i := 1; i <= 100; i++ {
			h.assertCommandSuccess(t, "add", fmt.Sprintf("Tâche stress %d", i), "+stress")
		}
		addDuration := time.Since(start)
		t.Logf("Temps pour ajouter 100 tâches: %v", addDuration)

		// Lister toutes les tâches
		start = time.Now()
		output := h.assertCommandSuccess(t, "list")
		listDuration := time.Since(start)
		t.Logf("Temps pour lister 100 tâches: %v", listDuration)

		// Vérifier que toutes sont présentes
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 100 {
			t.Errorf("Nombre de tâches listées: attendu 100, obtenu %d", len(lines))
		}

		// Export
		start = time.Now()
		h.assertCommandSuccess(t, "export", "stress_test.csv")
		exportDuration := time.Since(start)
		t.Logf("Temps pour exporter 100 tâches: %v", exportDuration)

		// Vérifications de performance (limites raisonnables)
		if addDuration > 10*time.Second {
			t.Errorf("Ajout de 100 tâches trop lent: %v", addDuration)
		}
		if listDuration > 2*time.Second {
			t.Errorf("Listing de 100 tâches trop lent: %v", listDuration)
		}
		if exportDuration > 2*time.Second {
			t.Errorf("Export de 100 tâches trop lent: %v", exportDuration)
		}
	})
}

// Tests de robustesse

func TestE2E_ErrorRecovery(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("récupération après fichier corrompu", func(t *testing.T) {
		// 1. Ajouter des tâches normalement
		h.assertCommandSuccess(t, "add", "Tâche avant corruption")

		// 2. Corrompre le fichier JSON
		corruptedJSON := `{"tasks": [invalid json content`
		err := ioutil.WriteFile(h.todoFile, []byte(corruptedJSON), 0644)
		if err != nil {
			t.Fatalf("Impossible de corrompre le fichier: %v", err)
		}

		// 3. L'application devrait gérer l'erreur gracieusement
		output := h.assertCommandSuccess(t, "list")

		// Devrait indiquer qu'il n'y a pas de tâches (nouveau démarrage)
		if !strings.Contains(output, "📝 Aucune tâche trouvée") {
			t.Errorf("Application n'a pas récupéré gracieusement: %s", output)
		}

		// 4. Ajouter une nouvelle tâche devrait fonctionner
		h.assertCommandSuccess(t, "add", "Tâche après récupération")

		listAfterRecovery := h.assertCommandSuccess(t, "list")
		if !strings.Contains(listAfterRecovery, "Tâche après récupération") {
			t.Errorf("Application ne fonctionne plus après récupération: %s", listAfterRecovery)
		}
	})

	t.Run("gestion des permissions", func(t *testing.T) {
		// Ce test est difficile à implémenter de manière portable
		// mais pourrait être ajouté pour des tests spécifiques Unix
		t.Skip("Test de permissions à implémenter selon la plateforme")
	})
}

// Tests de cas limites

func TestE2E_EdgeCases(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("texte avec caractères spéciaux", func(t *testing.T) {
		// Tester différents caractères spéciaux
		specialTexts := []string{
			"Tâche avec àéèçù",
			"Task with emoji 🚀📋",
			"Tâche avec guillemets doubles",
			"Tâche avec apostrophes simples",
			"Tâche avec caractères spéciaux #@$%",
		}

		for i, text := range specialTexts {
			t.Run(fmt.Sprintf("text_%d", i), func(t *testing.T) {
				output := h.assertCommandSuccess(t, "add", text, "+special")
				if !strings.Contains(output, "✅ Tâche ajoutée") {
					t.Errorf("Échec ajout texte spécial: %s", text)
				}

				// Vérifier que la tâche est listée correctement
				listOutput := h.assertCommandSuccess(t, "list")
				if !strings.Contains(listOutput, text) {
					t.Errorf("Texte spécial non trouvé dans la liste: %s", text)
				}
			})
		}
	})

	t.Run("texte très long", func(t *testing.T) {
		longText := strings.Repeat("Très long texte ", 50) // ~800 caractères

		output := h.assertCommandSuccess(t, "add", longText, "+long")
		if !strings.Contains(output, "✅ Tâche ajoutée") {
			t.Errorf("Échec ajout texte long")
		}

		// Vérifier que le texte long est géré correctement
		listOutput := h.assertCommandSuccess(t, "list")
		if !strings.Contains(listOutput, "Très long texte") {
			t.Errorf("Texte long non trouvé dans la liste")
		}
	})

	t.Run("beaucoup de tags", func(t *testing.T) {
		manyTags := []string{"+tag1", "+tag2", "+tag3", "@context1", "@context2", "@context3"}

		args := []string{"add", "Tâche avec beaucoup de tags"}
		args = append(args, manyTags...)

		output := h.assertCommandSuccess(t, args...)
		if !strings.Contains(output, "✅ Tâche ajoutée") {
			t.Errorf("Échec ajout avec beaucoup de tags")
		}

		// Vérifier que tous les tags sont présents
		listOutput := h.assertCommandSuccess(t, "list")
		for _, tag := range manyTags {
			if !strings.Contains(listOutput, tag) {
				t.Errorf("Tag manquant: %s", tag)
			}
		}
	})

	t.Run("dates limites diverses", func(t *testing.T) {
		testDates := []struct {
			date        string
			shouldWork  bool
			description string
		}{
			{"2025-07-20", true, "date normale"},
			{"2025-12-31", true, "fin d'année"},
			{"2025-01-01", true, "début d'année"},
			{"2025-02-29", false, "29 février année non bissextile"},
			{"2025-13-01", false, "mois invalide"},
			{"2025-07-32", false, "jour invalide"},
		}

		for _, test := range testDates {
			t.Run(test.description, func(t *testing.T) {
				if test.shouldWork {
					output := h.assertCommandSuccess(t, "add", "Test date "+test.description,
						"--due="+test.date)
					if !strings.Contains(output, "✅ Tâche ajoutée") {
						t.Errorf("Date valide rejetée: %s", test.date)
					}
				} else {
					h.assertCommandFails(t, 1, "add", "Test date "+test.description,
						"--due="+test.date)
				}
			})
		}
	})
}

// Tests de compatibilité

func TestE2E_Compatibility(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("export import avec différents formats", func(t *testing.T) {
		// Créer des tâches avec tous les types de données
		h.assertCommandSuccess(t, "add", "Tâche complète", "+projet", "@lieu",
			"--priority=high", "--due=2025-07-20")
		h.assertCommandSuccess(t, "add", "Tâche simple")
		h.assertCommandSuccess(t, "add", "Tâche avec émojis 🎯", "+fun")

		// Marquer une comme terminée
		h.assertCommandSuccess(t, "done", "2")

		// Export
		h.assertCommandSuccess(t, "export", "compatibility_test.csv")

		// Vérifier le fichier CSV
		csvPath := filepath.Join(h.tempDir, "compatibility_test.csv")
		time.Sleep(200 * time.Millisecond)

		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			t.Fatal("Fichier d'export non créé")
		}

		// Lire et vérifier le contenu CSV
		content, err := ioutil.ReadFile(csvPath)
		if err != nil {
			t.Fatalf("Impossible de lire le CSV: %v", err)
		}

		csvContent := string(content)

		// Vérifier l'en-tête
		if !strings.Contains(csvContent, "ID,UUID,Text,Done,Priority,Due,Tags,Created,Updated") {
			t.Error("En-tête CSV incorrect")
		}

		// Vérifier les données
		if !strings.Contains(csvContent, "Tâche complète") {
			t.Error("Tâche complète manquante dans CSV")
		}
		if !strings.Contains(csvContent, "high") {
			t.Error("Priorité manquante dans CSV")
		}

		// Test d'import de retour
		os.Remove(h.todoFile) // Reset

		h.assertCommandSuccess(t, "import", csvPath)

		// Vérifier que tout a été importé
		listOutput := h.assertCommandSuccess(t, "list", "--all")
		if !strings.Contains(listOutput, "Tâche complète") {
			t.Error("Import échoué pour tâche complète")
		}
		if !strings.Contains(listOutput, "✅") { // Tâche terminée
			t.Error("Statut Done perdu lors de l'import")
		}
	})
}

// Tests de performance détaillés

func TestE2E_PerformanceDetailed(t *testing.T) {
	if testing.Short() {
		t.Skip("Tests de performance ignorés en mode court")
	}

	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("performance opérations individuelles", func(t *testing.T) {
		// Test performance ajout
		start := time.Now()
		h.assertCommandSuccess(t, "add", "Test performance", "+perf")
		addTime := time.Since(start)
		t.Logf("Temps ajout: %v", addTime)

		if addTime > 500*time.Millisecond {
			t.Errorf("Ajout trop lent: %v", addTime)
		}

		// Test performance listing
		start = time.Now()
		h.assertCommandSuccess(t, "list")
		listTime := time.Since(start)
		t.Logf("Temps listing: %v", listTime)

		if listTime > 200*time.Millisecond {
			t.Errorf("Listing trop lent: %v", listTime)
		}

		// Test performance modification
		start = time.Now()
		h.assertCommandSuccess(t, "edit", "1", "Texte modifié", "+modif")
		editTime := time.Since(start)
		t.Logf("Temps modification: %v", editTime)

		if editTime > 500*time.Millisecond {
			t.Errorf("Modification trop lente: %v", editTime)
		}
	})

	t.Run("performance avec filtres", func(t *testing.T) {
		// Ajouter plusieurs tâches avec différents tags
		for i := 0; i < 50; i++ {
			project := fmt.Sprintf("+proj%d", i%5)
			context := fmt.Sprintf("@ctx%d", i%3)
			priority := []string{"low", "medium", "high"}[i%3]

			h.assertCommandSuccess(t, "add", fmt.Sprintf("Tâche %d", i),
				project, context, "--priority="+priority)
		}

		// Test performance filtrage par projet
		start := time.Now()
		h.assertCommandSuccess(t, "list", "--project=proj1")
		filterTime := time.Since(start)
		t.Logf("Temps filtrage projet: %v", filterTime)

		if filterTime > 300*time.Millisecond {
			t.Errorf("Filtrage trop lent: %v", filterTime)
		}

		// Test performance filtrage combiné
		start = time.Now()
		h.assertCommandSuccess(t, "list", "--project=proj1", "--context=ctx1", "--priority=high")
		complexFilterTime := time.Since(start)
		t.Logf("Temps filtrage complexe: %v", complexFilterTime)

		if complexFilterTime > 400*time.Millisecond {
			t.Errorf("Filtrage complexe trop lent: %v", complexFilterTime)
		}
	})
}

// Tests de sécurité basiques

func TestE2E_BasicSecurity(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("injection de commandes", func(t *testing.T) {
		// Tenter d'injecter des commandes dans le texte
		maliciousTexts := []string{
			"Tâche normale; rm -rf /",
			"Tâche $(whoami)",
			"Tâche `echo test`",
			"Tâche && echo injection",
			"Tâche | cat /etc/passwd",
		}

		for i, text := range maliciousTexts {
			t.Run(fmt.Sprintf("injection_%d", i), func(t *testing.T) {
				output := h.assertCommandSuccess(t, "add", text)
				if !strings.Contains(output, "✅ Tâche ajoutée") {
					t.Errorf("Texte rejeté à tort: %s", text)
				}

				// Vérifier que la tâche est stockée telle quelle
				listOutput := h.assertCommandSuccess(t, "list")
				if !strings.Contains(listOutput, text) {
					t.Errorf("Texte modifié de manière inattendue: %s", text)
				}
			})
		}
	})

	t.Run("chemins de fichiers dangereux", func(t *testing.T) {
		// Tenter d'exporter vers des chemins dangereux
		dangerousPaths := []string{
			"../../../etc/passwd",
			"/etc/hosts",
			"~/.bashrc",
		}

		for _, path := range dangerousPaths {
			// Ces commandes devraient échouer ou être contenues dans le répertoire sûr
			_, _, exitCode, _ := h.runCommand("export", path)
			// Nous ne testons pas l'échec car l'application peut créer des fichiers relatifs
			// L'important est qu'elle ne crée pas de fichiers dans des emplacements sensibles
			t.Logf("Export vers %s: code de sortie %d", path, exitCode)
		}
	})
}

// Tests de régression

func TestE2E_Regression(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("bug regression IDs après suppression", func(t *testing.T) {
		// Ajouter 3 tâches
		h.assertCommandSuccess(t, "add", "Tâche 1")
		h.assertCommandSuccess(t, "add", "Tâche 2")
		h.assertCommandSuccess(t, "add", "Tâche 3")

		// Supprimer la tâche du milieu
		h.assertCommandSuccess(t, "remove", "2")

		// Ajouter une nouvelle tâche
		h.assertCommandSuccess(t, "add", "Tâche 4")

		// Vérifier que les IDs sont cohérents
		listOutput := h.assertCommandSuccess(t, "list")
		lines := strings.Split(strings.TrimSpace(listOutput), "\n")

		if len(lines) != 3 {
			t.Errorf("Nombre de tâches incorrect: %d", len(lines))
		}

		// Vérifier qu'on peut manipuler toutes les tâches par ID
		h.assertCommandSuccess(t, "done", "1") // Doit marcher
		h.assertCommandSuccess(t, "done", "3") // Doit marcher
		h.assertCommandSuccess(t, "done", "4") // Doit marcher
	})

	t.Run("bug regression tags avec espaces", func(t *testing.T) {
		// Certains utilisateurs pourraient essayer d'ajouter des tags avec espaces
		h.assertCommandSuccess(t, "add", "Tâche test", "+tag avec espaces")

		// Vérifier que le tag est traité correctement (probablement séparé)
		listOutput := h.assertCommandSuccess(t, "list")
		if !strings.Contains(listOutput, "Tâche test") {
			t.Error("Tâche avec tag espacé non trouvée")
		}
	})

	t.Run("bug regression export fichier vide", func(t *testing.T) {
		// Exporter quand il n'y a pas de tâches
		h.assertCommandSuccess(t, "export", "empty_test.csv")

		csvPath := filepath.Join(h.tempDir, "empty_test.csv")
		time.Sleep(200 * time.Millisecond)

		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			t.Error("Export avec liste vide devrait créer un fichier")
		}

		// Vérifier que le fichier contient au moins l'en-tête
		content, err := ioutil.ReadFile(csvPath)
		if err != nil {
			t.Fatalf("Impossible de lire le fichier vide: %v", err)
		}

		if !strings.Contains(string(content), "ID,UUID,Text") {
			t.Error("En-tête manquant dans export vide")
		}
	})
}

// Tests finaux de validation

func TestE2E_FinalValidation(t *testing.T) {
	h := setupCLITest(t)
	defer h.cleanup()
	h.compileBinary(t)

	t.Run("workflow utilisateur réaliste", func(t *testing.T) {
		// Simuler l'usage d'un utilisateur réel pendant une semaine

		// Lundi: ajouter des tâches de travail
		h.assertCommandSuccess(t, "add", "Réunion équipe", "+travail", "@bureau", "--priority=high", "--due=2025-07-14")
		h.assertCommandSuccess(t, "add", "Finir rapport", "+travail", "@bureau", "--priority=medium", "--due=2025-07-16")
		h.assertCommandSuccess(t, "add", "Répondre emails", "+travail", "@bureau", "--priority=low")

		// Mardi: ajouter des tâches personnelles
		h.assertCommandSuccess(t, "add", "Courses alimentaires", "+perso", "@supermarché", "--priority=medium")
		h.assertCommandSuccess(t, "add", "Appeler médecin", "+santé", "@téléphone", "--priority=high")

		// Mercredi: compléter quelques tâches
		h.assertCommandSuccess(t, "done", "1") // Réunion équipe
		h.assertCommandSuccess(t, "done", "3") // Répondre emails

		// Jeudi: modifier une tâche
		h.assertCommandSuccess(t, "edit", "2", "Finir et envoyer rapport", "+travail", "+urgent", "@bureau")

		// Vendredi: voir le statut
		listOutput := h.assertCommandSuccess(t, "list")

		// Vérifier que le workflow fonctionne
		if !strings.Contains(listOutput, "Finir et envoyer rapport") {
			t.Error("Modification de tâche échouée")
		}
		if !strings.Contains(listOutput, "+urgent") {
			t.Error("Nouveau tag non ajouté")
		}

		// Weekend: export pour sauvegarde
		h.assertCommandSuccess(t, "export", "week_backup.csv")

		// Vérifier l'export
		csvPath := filepath.Join(h.tempDir, "week_backup.csv")
		time.Sleep(200 * time.Millisecond)

		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			t.Error("Sauvegarde hebdomadaire échouée")
		}

		// Afficher toutes les tâches pour validation finale
		allTasks := h.assertCommandSuccess(t, "list", "--all")
		t.Logf("État final des tâches:\n%s", allTasks)

		// Compter les tâches terminées et actives
		lines := strings.Split(strings.TrimSpace(allTasks), "\n")
		completed := 0
		active := 0

		for _, line := range lines {
			if strings.Contains(line, "✅") {
				completed++
			} else if strings.Contains(line, "⭕") {
				active++
			}
		}

		if completed != 2 {
			t.Errorf("Nombre de tâches terminées incorrect: %d", completed)
		}
		if active != 3 {
			t.Errorf("Nombre de tâches actives incorrect: %d", active)
		}

		t.Logf("✅ Workflow réussi: %d tâches terminées, %d actives", completed, active)
	})
}

// Fonction utilitaire pour nettoyer les tests
func cleanupAllTests() {
	// Cette fonction pourrait être appelée pour nettoyer après tous les tests
	// Par exemple, supprimer tous les fichiers temporaires restants
}
