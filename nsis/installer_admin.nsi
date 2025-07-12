; Todo CLI Installer - Version Admin ultra-simple
; Installation système avec droits administrateur

!define APPNAME "Todo CLI"
!define COMPANYNAME "NDXDeveloper"
!define DESCRIPTION "Gestionnaire de tâches en ligne de commande"
!define VERSIONMAJOR 0
!define VERSIONMINOR 0
!define VERSIONBUILD 9

; Configuration de l'installateur
Name "${APPNAME}"
OutFile "todo-setup-admin-v${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}.exe"

; Installation dans Program Files (nécessite admin)
InstallDir "$PROGRAMFILES64\${COMPANYNAME}\${APPNAME}"

; Droits admin requis
RequestExecutionLevel admin

; Métadonnées de l'installateur
VIProductVersion "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}.0"
VIAddVersionKey "ProductName" "${APPNAME}"
VIAddVersionKey "CompanyName" "${COMPANYNAME}"
VIAddVersionKey "FileDescription" "${DESCRIPTION}"
VIAddVersionKey "FileVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}"
VIAddVersionKey "ProductVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}"
VIAddVersionKey "LegalCopyright" "© ${COMPANYNAME}"

; Pages de l'installateur
Page directory
Page instfiles

; Section d'installation principal
Section "install"
    ; Créer le répertoire d'installation
    CreateDirectory "$INSTDIR"

    ; Copier les fichiers
    SetOutPath "$INSTDIR"
    File "todo-windows-amd64.exe"
    File /oname=todo.exe "todo-windows-amd64.exe"

    ; Ajouter au PATH système - méthode brutale mais qui marche
    DetailPrint "Configuration du PATH système..."
    ReadRegStr $0 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH"

    ; Ajouter notre chemin (même s'il existe déjà, pas grave)
    StrCmp $0 "" 0 +3
    WriteRegStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH" "$INSTDIR"
    Goto PathDone
    WriteRegStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PATH" "$0;$INSTDIR"

    PathDone:
    ; Notifier Windows du changement
    SendMessage 0xFFFF 0x001A 0 "STR:Environment" /TIMEOUT=5000
    DetailPrint "PATH système mis à jour"

    ; Créer l'entrée de désinstallation
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayName" "${APPNAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "UninstallString" "$INSTDIR\uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "Publisher" "${COMPANYNAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayIcon" "$INSTDIR\todo.exe"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoRepair" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "EstimatedSize" 3072

    ; Créer le désinstallateur
    WriteUninstaller "$INSTDIR\uninstall.exe"

    DetailPrint "Installation terminée avec succès !"

    ; Message de fin
    MessageBox MB_YESNO|MB_ICONQUESTION "Installation terminée !$\n$\nOuvrir un terminal pour tester 'todo' ?$\n$\nLa commande sera disponible dans tous les nouveaux terminaux." IDNO +2
    ExecShell "open" "cmd" "/k set PATH=%PATH%;$INSTDIR && echo Todo CLI installe ! && todo --help"
SectionEnd

; Section de désinstallation
Section "uninstall"
    ; Note: On ne nettoie pas le PATH système pour éviter les erreurs
    ; L'utilisateur peut le faire manuellement si nécessaire
    DetailPrint "Nettoyage des fichiers..."

    ; Supprimer les fichiers
    Delete "$INSTDIR\todo.exe"
    Delete "$INSTDIR\todo-windows-amd64.exe"
    Delete "$INSTDIR\uninstall.exe"

    ; Supprimer le répertoire d'installation
    RMDir "$INSTDIR"

    ; Supprimer l'entrée de désinstallation
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}"

    DetailPrint "Désinstallation terminée !"
    MessageBox MB_ICONINFORMATION "Désinstallation terminée !$\n$\nNote: Le PATH système n'a pas été modifié.$\nVous pouvez le nettoyer manuellement si souhaité."
SectionEnd
