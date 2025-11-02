# ShackoDodo
Hackathon A2025

## Build et Exécution

### Option 1: Exécutable unique avec interface intégrée (Recommandé)

1. **Build l'application complète** (Frontend React + Backend Go):
   ```bash
   build.bat
   ```

2. **Lancer l'application**:
   ```bash
   run.bat
   ```
   
   Ou directement:
   ```bash
   cd shack-o-hunter
   ShackoDodo.exe
   ```

L'application va:
- Démarrer le proxy sur le port 8181
- Démarrer le WebSocket sur le port 8182
- Servir l'interface React intégrée sur http://localhost:3000
- Ouvrir automatiquement le navigateur

### Option 2: Compilation rapide du backend seul

Si vous voulez juste compiler le backend Go sans le frontend:

```bash
cd shack-o-hunter
go build -o ShackoDodo.exe main.go
```

Dans ce mode, le proxy et le WebSocket fonctionneront, mais l'interface web ne sera pas disponible.
Utilisez `payload-modifier.html` et `browser-launcher.html` à la place.

### Option 3: Développement séparé

**Frontend (React + Vite):**
```bash
cd shack-o-dream
npm install
npm run dev
```

**Backend (Go):**
```bash
cd shack-o-hunter
go run main.go
```

## Architecture

- **shack-o-dream/**: Frontend React avec Material-UI
- **shack-o-hunter/**: Backend Go avec proxy HTTP/HTTPS
  - Proxy intercepteur sur port 8181
  - WebSocket sur port 8182
  - Serveur HTTP intégré pour le frontend (port 3000)

## Fonctionnalités

- Interception et modification des requêtes HTTP/HTTPS
- Interface web pour visualiser et éditer les requêtes
- Lancement de navigateurs avec proxy configuré
- Support Firefox, Chrome et Edge
- Installation automatique des certificats CA

