@echo off
echo ==========================================
echo Building ShackoDodo - Single Executable
echo ==========================================
echo.

REM Aller dans le dossier React
cd shack-o-dream
echo [1/4] Installation des dependances npm...
call npm install
if %errorlevel% neq 0 (
    echo Erreur lors de l'installation des dependances npm
    exit /b %errorlevel%
)

echo.
echo [2/4] Build du frontend React...
call npm run build
if %errorlevel% neq 0 (
    echo Erreur lors du build du frontend
    exit /b %errorlevel%
)

echo.
echo [3/4] Copie des fichiers build vers le serveur Go...
REM Supprimer l'ancien dossier dist s'il existe
if exist "..\shack-o-hunter\server\dist" rmdir /s /q "..\shack-o-hunter\server\dist"
REM Copier le nouveau dossier dist
xcopy /E /I /Y "dist" "..\shack-o-hunter\server\dist"
if %errorlevel% neq 0 (
    echo Erreur lors de la copie des fichiers
    exit /b %errorlevel%
)

REM Retourner à la racine
cd ..

echo.
echo [4/4] Compilation de l'executable Go...
cd shack-o-hunter
go build -o ShackoDodo.exe main.go
if %errorlevel% neq 0 (
    echo Erreur lors de la compilation Go
    exit /b %errorlevel%
)

REM Retourner à la racine
cd ..

echo.
echo ==========================================
echo Build terminé avec succes!
echo Executable: shack-o-hunter\ShackoDodo.exe
echo ==========================================
echo.
echo Pour lancer: cd shack-o-hunter ^&^& ShackoDodo.exe

