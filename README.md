# ShackoDodo
Hackathon A2025

## Build et Exécution

### Option 1: Exécutable unique (Recommandé)

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
- Servir l'interface React sur http://localhost:3000
- Ouvrir automatiquement le navigateur

### Option 2: Développement séparé

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

