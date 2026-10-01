# 🛒 Projet E-Commerce CLI (Go)

Une application d'E-Commerce complète et interactive en ligne de commande (TUI - Terminal User Interface) développée en Go.
Le projet respecte une architecture Backend moderne (API REST) couplée à deux interfaces clientes distinctes (Client & Administrateur) accessibles directement dans le terminal ou via une connexion SSH.

---

## ✨ Fonctionnalités Principales

### 🧑‍💻 Côté Client (Acheteur)
- **Authentification sécurisée** : Inscription avec hachage de mot de passe, validation de compte par code, connexion via JWT, et réinitialisation de mot de passe.
- **Catalogue & Recherche** : Parcourir les produits disponibles et recherche floue (full-text search) via l'API.
- **Panier d'achats** : Ajout et gestion de produits dans un panier avec persistance en base de données.
- **Paiement Stripe** : Intégration de l'API Stripe pour la validation et le paiement par carte bancaire.
- **Historique des commandes** : Consultation des commandes passées avec le détail des produits et suivi du statut (En préparation, expédié, livré).

### 🛡️ Côté Administrateur (Gérant)
- **Gestion des produits** : Ajout, modification, ou suppression de produits dans le catalogue.
- **Gestion des commandes** : Validation des paiements et changement de statut d'une commande (ex: passer de "Payé" à "Expédié").
- **Gestion des utilisateurs** : Droit de modification et de bannissement/suppression d'utilisateurs.

---

## 🏗️ Architecture & Technologies

- **Langage** : Go (Golang)
- **Base de données** : PostgreSQL 16
- **Déploiement / Conteneurisation** : Docker & Docker Compose
- **Interface TUI** : [Charmbracelet](https://charm.sh/) (`bubbletea`, `huh`, `lipgloss`) pour une interface terminal riche, vivante et interactive.
- **SSH Cloud TUI** : [Wish](https://github.com/charmbracelet/wish) pour rendre l'interface terminal nativement accessible via une connexion SSH réseau.
- **Sécurité** : JWT (JSON Web Tokens), `bcrypt` (hachage des mots de passe), et variables d'environnement (`.env`).

---

## 🚀 Guide de Démarrage

### 1. Prérequis
- [Go](https://go.dev/dl/) (version 1.20 ou supérieure)
- [Docker & Docker Compose](https://www.docker.com/)
- Un compte Stripe (Clé Secrète de test)

### 2. Installation
Clonez le dépôt et préparez les variables d'environnement en copiant le fichier d'exemple :
```bash
git clone https://github.com/ahmataffadine99/Go.git
cd Go
cp .env.example .env
```
*(N'oubliez pas d'éditer le fichier `.env` pour y ajouter votre clé secrète Stripe `STRIPE_SECRET_KEY`)*.

### 3. Lancement du Backend & Base de données
Utilisez Docker pour lever l'architecture serveur (Base de données PostgreSQL + API Go) en arrière-plan :
```bash
docker-compose up --build -d
```
Le serveur API tourne désormais sur `http://localhost:8080`.

---

## 🎮 Utilisation de l'Application

Vous avez deux manières de lancer les applications clientes (Client ou Admin) :

### Option A : Exécution Native Locale
Lancer directement l'application interactive dans votre propre terminal :
```bash
# Lancer l'interface Client (Acheteurs)
go run cmd/client/main.go

# Lancer l'interface Admin (Gérants)
go run cmd/admin/main.go
```

### Option B : Connexion via Réseau SSH (Cloud TUI)
Pour permettre à n'importe qui de se connecter au magasin sans installer l'application, lancez le serveur SSH :
```bash
# Lancer le serveur SSH (écoute sur le port 2222)
go run cmd/ssh/main.go
```
Puis, dans un autre terminal (ou depuis un autre ordinateur du réseau), connectez-vous :
```bash
ssh localhost -p 2222
```
*L'interface s'affichera directement au travers du tunnel SSH de manière magique !*

---
*Projet universitaire réalisé dans le cadre de l'apprentissage du développement Backend et Système en langage Go.*

**Développeurs :**
- AHMAT ABDOULAYE AFFADINE
- Elwardi Abderazzakh
