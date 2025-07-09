# 📋 Todo CLI Go

> 🚀 Gestionnaire de tâches moderne en ligne de commande écrit en Go - Rapide, portable et multiplateforme

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=for-the-badge)](https://github.com/NDXDev/todo-cli-go)
[![Release](https://img.shields.io/github/v/release/NDXDev/todo-cli-go?style=for-the-badge)](https://github.com/NDXDev/todo-cli-go/releases)

## 🎯 Aperçu

**Todo CLI Go** est un gestionnaire de tâches puissant et léger pour la ligne de commande. Développé en Go, il offre des performances exceptionnelles et une portabilité totale sur Windows, Linux et macOS.

### ✨ Fonctionnalités principales

- 🚀 **Ultra-rapide** : Écrit en Go pour des performances optimales
- 🎨 **Interface colorée** : Priorités visuelles et statuts clairs
- 📅 **Gestion des dates** : Dates limites avec alertes visuelles
- 🏷️ **Tags intelligents** : Projets (`+dev`) et contextes (`@bureau`)
- 🔍 **Filtrage avancé** : Par projet, priorité, statut
- 📊 **Export/Import CSV** : Sauvegarde et synchronisation
- 🔗 **UUID unique** : Import/export robuste sans conflits
- 🌐 **Multiplateforme** : Un seul binaire pour tous les OS
- 💾 **Stockage JSON** : Format lisible et portable

## 🖼️ Aperçu visuel

```
[1] ⭕ ❗ [due:2025-07-15] Préparer CV pour xxx@gmail.com +job @maison
[2] ✅ Envoyer le bilan France Travail +admin @maison [done:2025-07-09 14:30:00]
[3] ⭕ ⚠️ [due:2025-07-20] Réviser Go +dev @maison
```

## 🚀 Installation rapide

### Option 1 : Télécharger le binaire

```bash
# Linux/macOS
curl -L https://github.com/NDXDev/todo-cli-go/releases/latest/download/todo-linux -o todo
chmod +x todo
sudo mv todo /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/NDXDev/todo-cli-go/releases/latest/download/todo-windows.exe" -OutFile "todo.exe"
```

### Option 2 : Compiler depuis les sources

```bash
# Cloner le dépôt
git clone https://github.com/NDXDev/todo-cli-go.git
cd todo-cli-go

# Compiler
go build -o todo *.go

# Installer globalement (optionnel)
sudo mv todo /usr/local/bin/
```

### Option 3 : Installation via Go

```bash
go install github.com/NDXDev/todo-cli-go@latest
```

## 📖 Guide d'utilisation

### Commandes de base

```bash
# Ajouter une tâche
todo add "Préparer présentation +travail @bureau"

# Ajouter avec priorité et date limite
todo add "Rendez-vous client" +vente @ville --priority=high --due=2025-07-20

# Lister les tâches
todo list

# Marquer comme terminée
todo done 1

# Supprimer une tâche
todo remove 2

# Modifier une tâche
todo edit 3 "Nouvelle description" +urgent @bureau
```

### Gestion des tags

**Tags séparés du texte** (recommandé) :
```bash
# Tags comme arguments séparés
todo add "Email à marie@entreprise.com" +travail @bureau --priority=high

# Texte libre + tags explicites
todo add "Calculer 2+2=4 pour le projet" +math @école
```

### Filtrage avancé

```bash
# Filtrer par projet
todo list --project=travail

# Filtrer par contexte
todo list --context=bureau

# Filtrer par priorité
todo list --priority=high

# Combiner les filtres
todo list --project=dev --context=maison --priority=medium

# Afficher toutes les tâches (y compris terminées)
todo list --all
```

### Export et Import CSV

```bash
# Exporter toutes les tâches
todo export mes_taches.csv

# Import simple (mode merge par défaut)
todo import backup.csv

# Import avec stratégies avancées
todo import external.csv --mode=merge --conflict=newer --verbose

# Aperçu sans modification
todo import tasks.csv --dry-run

# Remplacement complet (destructif)
todo import full_backup.csv --mode=replace
```

## 🔧 Options et paramètres

### Options pour `add`
| Option | Alias | Description |
|--------|-------|-------------|
| `--priority` | `-p` | Priorité (low, medium, high) |
| `--due` | `-d` | Date limite (YYYY-MM-DD) |

### Options pour `list`
| Option | Alias | Description |
|--------|-------|-------------|
| `--all` | `-a` | Afficher toutes les tâches |
| `--project` | | Filtrer par projet (+tag) |
| `--context` | | Filtrer par contexte (@tag) |
| `--priority` | | Filtrer par priorité |

### Options pour `import`
| Option | Description | Valeurs |
|--------|-------------|---------|
| `--mode` | Mode d'import | `merge` (défaut), `replace` |
| `--conflict` | Stratégie de conflit | `skip` (défaut), `update`, `newer` |
| `--dry-run` | Aperçu sans modification | |
| `--verbose` | Mode détaillé | |

### Système de tags

- **Projets** : `+dev`, `+travail`, `+perso`
- **Contextes** : `@bureau`, `@maison`, `@ville`
- **Exemple** : `"Coder nouvelle feature" +dev @bureau`

### Priorités visuelles

- 🔴 **Haute** (`high`) : ❗ rouge
- 🟡 **Moyenne** (`medium`) : ⚠️ jaune
- 🔵 **Basse** (`low`) : ℹ️ bleu

## 🔧 Configuration

### Stockage des données

Les tâches sont sauvegardées dans :
- **Linux/macOS** : `$HOME/.todo/todo.json`
- **Windows** : `%USERPROFILE%\.todo\todo.json`

### Format JSON

```json
{
  "tasks": [
    {
      "id": 1,
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "text": "Préparer CV pour xxx@gmail.com",
      "done": false,
      "priority": "high",
      "due": "2025-07-15",
      "tags": ["+job", "@maison"],
      "created": "2025-07-09 14:30:00",
      "updated": "2025-07-09 14:30:00"
    }
  ],
  "nextId": 2
}
```

### Autocomplétion Bash

Ajoutez à votre `~/.bashrc` :

```bash
_todo_completion() {
    local cur prev
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "add list done remove edit export import help" -- $cur) )
    fi
}
complete -F _todo_completion todo
```

## 🧪 Tests et exemples

### Scénario complet

```bash
# 1. Ajouter des tâches variées
todo add "Réunion équipe" +travail @bureau --priority=high --due=2025-07-15
todo add "Courses hebdomadaires" +perso @supermarché --priority=medium
todo add "Réviser Go pour certification" +dev @maison --priority=low

# 2. Visualiser et filtrer
todo list
todo list --project=travail --priority=high
todo list --context=maison

# 3. Compléter des tâches
todo done 1

# 4. Modifier une tâche
todo edit 2 "Courses + pharmacie" +perso @centre-ville

# 5. Export et sauvegarde
todo export rapport_hebdo.csv

# 6. Import depuis autre source
todo import team_tasks.csv --mode=merge --conflict=newer --verbose
```

### Test de cycle export/import

```bash
# Créer des tâches de test
todo add "Tâche 1" +test @local --priority=high
todo add "Tâche 2" +test @remote --priority=medium

# Export
todo export test_backup.csv

# Simuler perte de données
rm ~/.todo/todo.json

# Import et restauration
todo import test_backup.csv --verbose

# Vérification
todo list
```

## 🛠️ Développement

### Prérequis

- Go 1.22+
- Git

### Structure du projet

```
todo-cli-go/
├── main.go             # Code principal et CLI
├── import.go           # Fonctions d'import CSV
├── README.md           # Documentation
├── LICENSE             # Licence MIT
├── go.mod              # Module Go
├── .gitignore          # Fichiers ignorés
└── releases/           # Binaires compilés
```

### Compilation pour différentes plateformes

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o todo-linux *.go

# Windows
GOOS=windows GOARCH=amd64 go build -o todo-windows.exe *.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o todo-macos *.go

# Architecture actuelle
go build -o todo *.go
```

### Tests

```bash
# Lancer les tests
go test ./...

# Tests avec couverture
go test -cover ./...

# Benchmark
go test -bench=. ./...
```

## 🤝 Contribution

Les contributions sont les bienvenues ! Voici comment participer :

1. **Fork** le projet
2. **Créer** une branche feature (`git checkout -b feature/AmazingFeature`)
3. **Commit** vos changements (`git commit -m 'Add: AmazingFeature'`)
4. **Push** vers la branche (`git push origin feature/AmazingFeature`)
5. **Ouvrir** une Pull Request

### Idées d'améliorations

- [ ] Synchronisation cloud (Google Drive, Dropbox)
- [ ] Interface web optionnelle
- [ ] Notifications desktop
- [ ] Import depuis d'autres formats (Todoist, Trello)
- [ ] Rapports et statistiques avancées
- [ ] Thèmes de couleurs personnalisables
- [ ] Récurrence de tâches
- [ ] Sous-tâches et dépendances

## 📝 Changelog

### v1.0.0 (2025-07-09)
- ✨ Première version stable
- 🚀 Toutes les fonctionnalités de base
- 🎨 Interface colorée avec priorités visuelles
- 📊 Export/Import CSV complet
- 🔍 Filtrage avancé par projet, contexte, priorité
- 🔗 UUID unique pour import/export robuste
- 🏷️ Système de tags séparés du texte
- 📅 Gestion des dates limites
- 🌐 Support multiplateforme complet

## 🐛 Signaler un bug

Si vous rencontrez un problème :

1. Vérifiez les [issues existantes](https://github.com/NDXDev/todo-cli-go/issues)
2. Créez une [nouvelle issue](https://github.com/NDXDev/todo-cli-go/issues/new) avec :
   - Description détaillée du problème
   - Commande exacte utilisée
   - Comportement attendu vs actuel
   - Environnement (OS, version Go)
   - Fichiers de log si applicable

## 📧 Contact

- **Développeur** : NDXDev
- **Email** : NDXDev@gmail.com
- **GitHub** : [@NDXDev](https://github.com/NDXDeveloper)

## 📄 Licence

Ce projet est sous licence MIT. Voir le fichier [LICENSE](LICENSE) pour plus de détails.

## 🌟 Remerciements

- La communauté Go pour l'excellent écosystème
- Les utilisateurs et contributeurs
- L'inspiration du format todo.txt et Getting Things Done (GTD)

---

⭐ **N'hésitez pas à mettre une étoile si ce projet vous plaît !**

[![GitHub stars](https://img.shields.io/github/stars/NDXDev/todo-cli-go?style=social)](https://github.com/NDXDev/todo-cli-go/stargazers)
