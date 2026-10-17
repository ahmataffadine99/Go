# Projet E-Commerce CLI (Go)

Une application d'E-Commerce complète et interactive en ligne de commande (TUI - Terminal User Interface) développée en Go.
Le projet respecte une architecture Backend moderne (API REST) couplée à deux interfaces clientes distinctes (Client & Administrateur) accessibles directement dans le terminal ou via une connexion SSH.

---

## Fonctionnalités Principales

### Côté Client (Acheteur)
- **Authentification sécurisée** : Inscription avec hachage salé du mot de passe (PBKDF2), validation de compte par code aléatoire, connexion par token signé (HMAC-SHA256) et réinitialisation du mot de passe par code envoyé par email (simulé dans les logs du serveur).
- **Catalogue & Recherche** : Recherche par nom, description, identifiant métier, catégorie, prix HT ou prix TTC (TVA 20%).
- **Panier d'achats** : Ajout, modification et suppression de produits dans un panier lié au compte (identifiant BSK-XXXXXX), total TTC avec livraison gratuite.
- **Paiement** : Saisie des informations bancaires (numéro de carte avec vérification par algorithme de Luhn, date d'expiration, CVC) avec simulation sécurisée du débit côté serveur.
- **Historique des commandes** : Consultation des commandes (identifiant CMD-XXXXXX) avec la date, le détail des produits et le statut (en attente, payé, en cours de livraison, livré, annulé avec la raison).

### Côté Administrateur (Gérant)
- **Gestion des produits** : Consultation du catalogue et ajout de produits avec génération de l'identifiant métier (PDT-XXXXXX).
- **Gestion des commandes** : Consultation de toutes les commandes, changement de statut (raison obligatoire pour une annulation) et création d'une commande pour le compte d'un client.
- **Gestion des utilisateurs** : Création, modification (email, rôle, mot de passe), confirmation et suppression de comptes.

---

## Architecture & Technologies

- **Langage** : Go (Golang)
- **Serveur HTTP** : module standard `net/http`, sans framework
- **Base de données** : `database/sql` avec PostgreSQL 16 (Docker) ou SQLite
- **Déploiement / Conteneurisation** : Docker & Docker Compose
- **Interface TUI** : [Charmbracelet](https://charm.sh/) (`bubbletea`, `bubbles`, `huh`, `lipgloss`) : formulaires, tableaux interactifs et indicateurs de chargement.
- **SSH Cloud TUI** : [Wish](https://github.com/charmbracelet/wish) pour rendre l'interface terminal nativement accessible via une connexion SSH réseau.
- **Sécurité** : tokens signés HMAC-SHA256, hachage PBKDF2 salé (bibliothèque standard) et variables d'environnement (`.env`).

---

## Guide de Démarrage

### 1. Prérequis
- [Go](https://go.dev/dl/) (version 1.24 ou supérieure)
- [Docker & Docker Compose](https://www.docker.com/)

### 2. Installation
Clonez le dépôt et préparez les variables d'environnement en copiant le fichier d'exemple :
```bash
git clone https://github.com/ahmataffadine99/Go.git
cd Go
cp .env.example .env
```

### 3. Lancement du Backend & Base de données
Utilisez Docker pour lever l'architecture serveur (Base de données PostgreSQL + API Go) en arrière-plan :
```bash
docker-compose up --build -d
```
Le serveur API tourne désormais sur `http://localhost:8080`.

Sans Docker, le serveur peut aussi être lancé avec SQLite (base créée dans `./ecommerce.db`) :
```bash
go run ./cmd/server
```

Au démarrage, un compte administrateur est créé s'il n'existe pas encore. Ses identifiants sont définis par `ADMIN_EMAIL` et `ADMIN_PASSWORD` (par défaut `admin@shop.local` / `admin123`).

> **Note sur la validation par email (Mock Email) :**  
> Conformément aux spécifications, l'envoi d'emails est simulé côté serveur. Les codes de confirmation d'inscription et de réinitialisation de mot de passe sont consultables en temps réel dans les logs du serveur :
> - Si le serveur tourne sous Docker : `docker logs -f ecommerce_backend`
> - Si le serveur tourne en local : les codes apparaissent directement dans le terminal exécutant `go run ./cmd/server`.

---

## Utilisation de l'Application

Vous avez deux manières de lancer les applications clientes (Client ou Admin) :

### Option A : Exécution Native Locale
Lancer directement l'application interactive dans votre propre terminal :
```bash
# Lancer l'interface Client (Acheteurs)
go run ./cmd/client
# ou explicitement en mode interactif
go run ./cmd/client interactive

# Lancer l'interface Admin (Gérants)
go run ./cmd/admin

# Commandes utilitaires du client CLI
go run ./cmd/client status     # Vérifie la connectivité avec l'API
go run ./cmd/client version    # Affiche la version du client

# Spécifier une adresse de serveur personnalisée
go run ./cmd/client -url http://localhost:8080
go run ./cmd/admin -url http://localhost:8080
```

### Option B : Connexion via Réseau SSH (Cloud TUI)
Pour permettre à n'importe qui de se connecter au magasin sans installer l'application sur son poste, lancez le serveur SSH :
```bash
# Lancer le serveur SSH (écoute par défaut sur le port 2222, configurable via SSH_PORT et API_URL)
go run ./cmd/ssh
```
Puis, dans un autre terminal (ou depuis une autre machine du réseau), connectez-vous :
```bash
# Accès à l'application client
ssh -p 2222 client@localhost

# Accès à l'application administrateur
ssh -p 2222 admin@localhost
```

### Tests Unitaires
Pour exécuter l'ensemble des tests automatisés :
```bash
go test ./... -v
```

---

## Routes de l'API

| Méthode | Route | Accès | Description |
|---|---|---|---|
| POST | `/api/auth/register` | public | Inscription |
| POST | `/api/auth/confirm` | public | Confirmation du compte |
| POST | `/api/auth/login` | public | Connexion |
| POST | `/api/auth/forgot-password` | public | Envoi d'un code de réinitialisation |
| POST | `/api/auth/reset-password` | public | Réinitialisation avec le code |
| GET | `/api/products` | public | Liste et recherche (`q`, `category`, `min_price`, `max_price`, `min_price_ttc`, `max_price_ttc`) |
| GET | `/api/cart` | client | Panier actif |
| POST | `/api/cart/add` | client | Ajout d'un produit |
| PUT / DELETE | `/api/cart/item?product_id=` | client | Modification ou suppression d'une ligne |
| POST | `/api/cart/pay` | client | Paiement du panier |
| GET | `/api/orders` | client | Commandes de l'utilisateur |
| POST | `/api/admin/products` | admin | Ajout d'un produit |
| GET | `/api/admin/orders` | admin | Toutes les commandes |
| PUT | `/api/admin/orders/status?id=` | admin | Changement de statut |
| POST | `/api/admin/orders/create` | admin | Création d'une commande pour un client |
| GET | `/api/admin/users` | admin | Liste des utilisateurs |
| POST | `/api/admin/users/create` | admin | Création d'un utilisateur |
| PUT | `/api/admin/users/update?id=` | admin | Modification d'un utilisateur |
| POST | `/api/admin/users/confirm?id=` | admin | Confirmation d'un compte |
| DELETE | `/api/admin/users/delete?id=` | admin | Suppression d'un utilisateur |

---
*Projet universitaire réalisé dans le cadre de l'apprentissage du développement Backend et Système en langage Go.*

**Développeurs :**
- AHMAT ABDOULAYE AFFADINE
- Elwardi Abderazzakh
