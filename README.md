Je participe
============

Solution simple pour organiser le "qui donne un coup de main à quoi" d'une manifestation en partant des souhaits de chacun.

Accessible à tous, cette solution est un simple exécutable qui fonctionne aussi bien sur un serveur ou directement sur le poste personnel d'un organisateur.

Fonctionnement
--------------

Chaque activité de l'évènement est associée à un "tableau des inscriptions" dans lequel les personnes qui souhaitent y prendre part inscrivent leur nom. 
Pour chaque activité disponible, il y a un nombre maximal de participants pouvant s'inscrire et un nombre souhaité de participants.

Exemple :
 * Description : "Tenue du bar de 14h à 16h"
 * Nombre souhaité de participants : 4
 * Nombre maximal de participants pouvant s'inscrire : 8
 
Petit à petit les "tableaux des inscriptions" pour chaque activité se remplissent. Leur fonctionnement est régi par les règles suivantes :
  1. N'importe qui peut inscrire le nom de n'importe qui dans un tableau d'inscription (pas de gestion de compte / pas de vérification)
  2. Dès qu'un nom est inscrit dans un tableau, celui qui l'a inscrit peut le supprimer tant que sa session est active (cookie de session d'une durée de vie par défaut à 1 heure)
  3. Toute personne peut demander la suppression de n'importe qui dans un tableau d'inscription, mais elle doit motiver sa décision. La suppression est effective qu'après accord des administrateurs
  5. A tout moment, un administrateur peut figer un tableau d'inscription pour une activité. Une fois un tableau d'inscription figé, seul l'administrateur peut en modifier le contenu. 

En pratique, tant qu'un tableau n'est pas figé, le tableau des inscriptions pour une activité peut être vu comme une "liste de postulants".

Cas pratique sur notre exemple :

Jacques, Henri, Pierre, Manuela, Jean, Hervé se sont inscrits pour tenir le bar de 14h à 16h. Il reste donc encore 2 personnes qui peuvent postuler pour cette activité. Mais en pratique, 4 personnes sont suffisantes.
Il se peut que ces 6 personnes est également postulées à d'autres activités à la même heure. Il revient à l'organisateur de décider l'activité retenue pour chacun.

Petit à petit, l'organisateur arbitre l'organisation finale en figeant les tableaux d'inscription de toutes les activités de la manifestation jusqu'à obtention de l'organisation finale.


Installation
------------

A décrire

Description des données
-----------------------

Lors du démarrage du serveur, la base de données est initialisée et le mot de passe administrateur est généré. Celui-ci est alors affiché dans la console.
A chaque redémarrage du serveur, le mot de passe administrateur est affiché dans la console pour rappel.

La base de données contient les informations suivantes :

* Mot de passe administrateur

* Demandes de suppression
  * nom du tableau d'inscription
  * nom de l'inscrit
  * motif du demandeur
  * horodatage de la demande
  * état (validé / refusé / non traité)
  * horodatage du traitement de la demande

* Tableaux des inscriptions :
  * nom du tableau d'inscription
  * etat du tableau (modifications ouvertes à tous ou limitées aux administrateurs)
  * liste des inscrits
