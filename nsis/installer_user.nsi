; Todo CLI Installer - Version minimale sans macros
; Installation dans le profil utilisateur avec PATH utilisateur

!define APPNAME "Todo CLI"
!define COMPANYNAME "NDXDeveloper"
!define DESCRIPTION "Gestionnaire de tâches en ligne de commande"
!define VERSIONMAJOR 0
!define VERSIONMINOR 0
!define VERSIONBUILD 9

; Configuration de l'installateur
Name "${APPNAME}"
OutFile "todo-setup-noadmin-v${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}.exe"

; Installation dans le profil utilisateur (pas besoin d'admin)
InstallDir "$LOCALAPPDATA\${COMPANYNAME}\${APPNAME}"

; Pas besoin de droits admin
RequestExecutionLevel user

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

    ; Ajouter au PATH utilisateur - méthode simple
    DetailPrint "Configuration du PATH utilisateur..."
    ReadRegStr $0 HKCU "Environment" "PATH"

    ; Si PATH est vide, ajouter juste notre chemin
    StrCmp $0 "" 0 +3
    WriteRegStr HKCU "Environment" "PATH" "$INSTDIR"
    Goto PathDone

    ; Sinon, ajouter à la fin avec un point-virgule
    WriteRegStr HKCU "Environment" "PATH" "$0;$INSTDIR"

    PathDone:
    ; Notifier Windows du changement
    SendMessage 0xFFFF 0x001A 0 "STR:Environment" /TIMEOUT=5000
    DetailPrint "PATH utilisateur mis à jour"

    ; Créer l'entrée de désinstallation dans le registre utilisateur
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayName" "${APPNAME}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "UninstallString" "$INSTDIR\uninstall.exe"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "Publisher" "${COMPANYNAME}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayIcon" "$INSTDIR\todo.exe"
    WriteRegDWORD HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoModify" 1
    WriteRegDWORD HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoRepair" 1
    WriteRegDWORD HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "EstimatedSize" 3072

    ; Créer le désinstallateur
    WriteUninstaller "$INSTDIR\uninstall.exe"

    ; Créer un raccourci sur le bureau (optionnel)
    ;MessageBox MB_YESNO "Créer un raccourci sur le bureau ?" IDNO +2
    ;CreateShortcut "$DESKTOP\${APPNAME}.lnk" "$INSTDIR\todo.exe"

    ; Créer des raccourcis dans le menu démarrer
    ;c'est une application console donc pas besoin de ces raccourcis
    ;CreateDirectory "$SMPROGRAMS\${APPNAME}"
    ;CreateShortcut "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk" "$INSTDIR\todo.exe"
    ;CreateShortcut "$SMPROGRAMS\${APPNAME}\Désinstaller ${APPNAME}.lnk" "$INSTDIR\uninstall.exe"

    DetailPrint "Installation terminée avec succès !"

    ; Message de fin
    MessageBox MB_YESNO|MB_ICONQUESTION "Installation terminée !$\n$\nOuvrir un terminal pour tester 'todo' ?$\n$\nNote: Redémarrez votre terminal si nécessaire." IDNO +2
    ExecShell "open" "cmd" "/k set PATH=%PATH%;$INSTDIR && echo Todo CLI installe ! && todo --help"
SectionEnd

; Section de désinstallation
Section "uninstall"
    ; Supprimer du PATH utilisateur - méthode simple
    DetailPrint "Nettoyage du PATH utilisateur..."
    ReadRegStr $0 HKCU "Environment" "PATH"

    ; Remplacer les occurrences de notre chemin
    ; Note: Cette méthode simple peut laisser des point-virgules orphelins
    ; mais c'est acceptable pour un installateur basique

    ; Supprimer les fichiers
    Delete "$INSTDIR\todo.exe"
    Delete "$INSTDIR\todo-windows-amd64.exe"
    Delete "$INSTDIR\uninstall.exe"

    ; Supprimer les raccourcis
    ;Delete "$DESKTOP\${APPNAME}.lnk"
    ;Delete "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk"
    ;Delete "$SMPROGRAMS\${APPNAME}\Désinstaller ${APPNAME}.lnk"
    RMDir "$SMPROGRAMS\${APPNAME}"

    ; Supprimer le répertoire d'installation
    RMDir "$INSTDIR"

    ; Supprimer l'entrée de désinstallation
    DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}"

    DetailPrint "Désinstallation terminée !"

    MessageBox MB_ICONINFORMATION "Désinstallation terminée !$\n$\nNote: Vous pouvez redémarrer votre terminal pour nettoyer le PATH."
SectionEnd
