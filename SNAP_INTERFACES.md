# Configuration des interfaces Snap pour Todo CLI Go

## 🔐 Interfaces disponibles

Le snap Todo CLI Go utilise un confinement **strict** avec les interfaces suivantes :

### Automatiquement connectées
- ✅ `home` - Accès au répertoire utilisateur (`~/`)
- ✅ `network` - Accès réseau de base

### Connexions manuelles requises

#### Pour l'accès aux clés USB et disques externes :
```bash
sudo snap connect todo-cli-go:removable-media
```

#### Pour l'observation des points de montage :
```bash
sudo snap connect todo-cli-go:mount-observe
```

#### Pour l'écoute réseau (si nécessaire) :
```bash
sudo snap connect todo-cli-go:network-bind
```

#### Pour l'accès aux partages réseau étendus :
```bash
sudo snap connect todo-cli-go:network-shares
```

## 🎯 Utilisation pratique

### Configuration complète pour tous les accès :
```bash
# Installation du snap
sudo snap install --dangerous todo-cli-go_v.x.x.x_amd64.snap

# Connexion de toutes les interfaces
sudo snap connect todo-cli-go:removable-media
sudo snap connect todo-cli-go:mount-observe
sudo snap connect todo-cli-go:network-bind

# Test d'accès
todo-cli-go version
```

### Vérification des connexions :
```bash
# Voir les interfaces connectées
snap connections todo-cli-go

# Exemple de sortie attendue :
# Interface        Plug                    Slot    Notes
# home             todo-cli-go:home        :home   -
# network          todo-cli-go:network     :network -
# removable-media  todo-cli-go:removable-media :removable-media manual
# mount-observe    todo-cli-go:mount-observe :mount-observe manual
```

## 📁 Accès aux répertoires

Avec les interfaces connectées, Todo CLI Go peut accéder à :

- **Répertoire home** : `~/` (automatique)
- **Clés USB** : `/media/username/`, `/mnt/` (avec removable-media)
- **Partages réseau montés** : Points de montage système (avec mount-observe)
- **GVFS** : `~/.gvfs/`, `/run/user/*/gvfs/` (partages GNOME/Ubuntu)

## 🔒 Sécurité

Le confinement strict garantit que :
- ❌ Aucun accès système non autorisé
- ✅ Permissions explicites et auditables
- ✅ Révocation facile des permissions
- ✅ Isolation complète du reste du système

## 🚀 Exemples d'usage

### Export vers clé USB :
```bash
# Après connexion de removable-media
todo export /media/username/USB_DRIVE/my_tasks.csv
```

### Import depuis partage réseau :
```bash
# Après montage du partage et connexion mount-observe
todo import /mnt/shared_drive/tasks_backup.csv
```
