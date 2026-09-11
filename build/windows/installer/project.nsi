Unicode true

####
## BSManager 用 NSIS インストーラー定義
##
## `wails build -nsis` で wails_tools.nsh が wails.json の info から自動生成され、
## 下記の INFO_* が未定義ならそこから補完されます。
## 手動でビルドする場合:
##   makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\BSManager.exe project.nsi
####

## 既定値を上書きしたい場合はここで定義する（wails_tools.nsh より前）
# !define INFO_PROJECTNAME    "BSManager"
# !define INFO_COMPANYNAME    "SayukiDev"
# !define INFO_PRODUCTNAME    "BSManager"
# !define INFO_PRODUCTVERSION "0.0.1"
# !define INFO_COPYRIGHT      "Copyright © 2026 SayukiDev"
# !define PRODUCT_EXECUTABLE  "BSManager.exe"
# !define UNINST_KEY_NAME     "SayukiBSManager"

## "admin": Program Files へ全ユーザー向けにインストール / "user": ユーザー単位（UAC なし）
!define REQUEST_EXECUTION_LEVEL "admin"

## wails_tools.nsh 内の英語メッセージを日本語化
!define WAILS_WIN10_REQUIRED "このアプリケーションは Windows 10 以降でのみ動作します。"
!define WAILS_ARCHITECTURE_NOT_SUPPORTED "このインストーラーは現在の Windows のアーキテクチャに対応していません。対応: ${ARCH}"
!define WAILS_INSTALL_WEBVIEW_DETAILPRINT "WebView2 ランタイムをインストールしています…"

!include "wails_tools.nsh"

## 設定ファイルの保存先（settings.DefaultDir と一致させること: %APPDATA%\BSManager）
!define APP_CONFIG_DIRNAME "BSManager"

# バージョン情報は 4 パート必須
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} セットアップ"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# HiDPI 対応 https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true
SetCompressor /SOLID lzma

!include "MUI2.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!define MUI_FINISHPAGE_RUN_TEXT "${INFO_PRODUCTNAME} を起動する"

# 起動中のアプリを検出したときのメッセージ用
!define MUI_UNABORTWARNING

!insertmacro MUI_PAGE_WELCOME
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "Japanese"

## 署名する場合はここを有効化
#!uninstfinalize 'signtool sign /fd sha256 /tr http://timestamp.digicert.com /td sha256 "%1"'
#!finalize 'signtool sign /fd sha256 /tr http://timestamp.digicert.com /td sha256 "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
InstallDirRegKey HKLM "${UNINST_KEY}" "InstallLocation"
ShowInstDetails show
ShowUninstDetails show

## 実行中の BSManager を終了させる（SingleInstanceLock のため上書きに失敗しないように）
!macro closeRunningApp
    nsExec::ExecToStack 'taskkill /IM "${PRODUCT_EXECUTABLE}" /F'
    Pop $0
    Pop $1
    Sleep 500
!macroend

Function .onInit
    !insertmacro wails.checkArchitecture
FunctionEnd

Section "-Install"
    !insertmacro wails.setShellContext

    !insertmacro closeRunningApp

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0
    CreateShortcut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}" "" "$INSTDIR\${PRODUCT_EXECUTABLE}" 0

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller

    SetRegView 64
    WriteRegStr HKLM "${UNINST_KEY}" "InstallLocation" "$INSTDIR"
    WriteRegDWORD HKLM "${UNINST_KEY}" "NoModify" 1
    WriteRegDWORD HKLM "${UNINST_KEY}" "NoRepair" 1
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    !insertmacro closeRunningApp

    # WebView2 のデータ
    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller

    # 設定とログはユーザーに確認してから削除する（サイレント時は残す）
    IfSilent done
    SetShellVarContext current
    IfFileExists "$AppData\${APP_CONFIG_DIRNAME}\*.*" 0 done
    MessageBox MB_YESNO|MB_ICONQUESTION "設定ファイルとログ（$AppData\${APP_CONFIG_DIRNAME}）も削除しますか？" IDNO done
    RMDir /r "$AppData\${APP_CONFIG_DIRNAME}"
    done:
SectionEnd
