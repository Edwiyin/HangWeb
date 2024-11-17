Conversation opened. 1 read message.

Skip to content
Using Gmail with screen readers
1 of 308
(no subject)
Inbox

Michel Mustafov
Attachments
2:01 PM (0 minutes ago)
to me


 One attachment
  •  Scanned by Gmail
# Guide de Configuration de HangmanWeb

Veuillez suivre ces étapes pour jouer au jeu. 

---

## Prérequis

Avant de commencer, assurez l'installation des outils suivants :

### 1. **Visual Studio Code (VS Code)**
-  [https://code.visualstudio.com/](https://code.visualstudio.com/).

### 2. **Golang**
  1. Ouvrez **VS Code**.
  2. Ouvrez le panneau **Extensions** 
  3. Recherchez l'extension **Go**.
  4. Cliquez sur **Installer**.

### 3. **Git Bash**  
- [https://git-scm.com/](https://git-scm.com/).

Pour vérifier que Git est bien installé, executez les commandes suivantes dans terminal: 
```bash
git --version
```

---


## Remarque sur les Routes

Toutes les routes définies dans cette application web distribuent des vues. Cela signifie que chaque route correspond
à une vue spécifique qui est générée par le serveur et envoyée au navigateur. Par exemple :
- La route `/` affiche la page d'accueil.
- La route `/game` affiche l'interface du jeu.
- La route `/scoreboard` affiche le résultat de la partie.

Ces vues permettent de structurer clairement les différentes étapes du jeu et d'améliorer l'expérience utilisateur.

## Étapes pour Configurer le Projet

### 1. Cloner le Dépôt
1. Ouvrez un terminal et naviguez vers le dossier où vous souhaitez enregistrer le projet.
2. Exécutez la commande suivante pour cloner le projet :
   ```bash
   git clone https://github.com/Edwiyin/HangWeb.git
   ```
3. Accédez au répertoire du projet :
   ```bash
   cd HangWeb
   ```

### 2. Lancer le Serveur
1. Ouvrez le dossier du projet dans VS Code :
   ```bash
   code .
   ```
2. Exécutez la commande suivante pour démarrer le serveur :
   ```bash
   go run main.go
   ```

---

## Accéder au Site Web

1. Ouvrez votre navigateur web.
2. Saisissez l'URL suivante et appuyez sur Entrée :
   ```
   http://localhost:8080
   ```

Vous devriez maintenant voir le jeu **HangmanWeb** fonctionner !

---

## Résolution de Problèmes

- En cas de problème, vérifiez que :
  - Golang et Git sont correctement installés en exécutant `go version` et `git --version`.
  - Le terminal pointe vers le bon dossier du projet.
  - Aucun autre service n'utilise le port `8080`.

## Projet réalisé par Edwin WEHBE et Michel MUSTAFOV, étudiants à Aix Ynov Campus. 


README.md
Displaying README.md.
