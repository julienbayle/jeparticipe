Je participe
============

Solution simple pour organiser le "qui donne un coup de main à quoi" d'une manifestation en partant des souhaits de chacun.

Accessible à tous, cette solution est un simple exécutable qui fonctionne aussi bien sur un serveur que directement sur le poste personnel d'un organisateur.

Fonctionnement
--------------

Chaque activité de la manifestation est associée à un "tableau des participants" dans lequel les personnes qui souhaitent y prendre part inscrivent leur nom. Au départ, il s'agit de recueillir les souhaits de participation, on peut donc parler de "tableau des postulants". Pour chaque tableau, il y a un nombre maximal de postulants pouvant s'inscrire et un nombre souhaité de participants.

Exemple :
 * Description : "Tenue du bar de 14h à 16h"
 * Nombre souhaité de participants : 4
 * Nombre maximal de postulants : 8
 
Petit à petit les "tableaux des participants" pour chaque activité se remplissent. Leur fonctionnement est régi par les règles suivantes :
  1. N'importe qui peut inscrire le nom de n'importe qui dans un tableau (pas besoin de se créer un compte)
  2. Toute inscription peut être librement annulée par celui qui l'a réalisé tant qu'il n'a pas fermé son navigateur web.
  3. N'importe qui peut demander l'annulation de participation de n'importe qui dans un tableau. Cette demande d'annulation est accompagnée d'une explication à destination des administrateurs. L'annulation est effective qu'après accord d'un administrateur
  4. Un administrateur peut figer un à un chaque "tableau de participants". Une fois un "tableau des participants" figé, seul un administrateur peut en modifier les participants.

En pratique, tant qu'un tableau n'est pas figé, le "tableau des participants" pour une activité peut être vu comme une "liste de postulants" pour cette activité.

Cas pratique sur notre exemple :

Jacques, Henri, Pierre, Manuela, Jean, Hervé se sont inscrits pour tenir le bar de 14h à 16h. Il reste donc encore 2 personnes qui peuvent postuler pour cette activité. Mais en pratique, 4 personnes sont suffisantes.
Il se peut que ces 6 personnes est également postulées à d'autres activités à la même heure. Il revient à l'organisateur de décider l'activité retenue pour chacun.

Petit à petit, l'organisateur arbitre l'organisation finale en figeant les tableaux de toutes les activités de la manifestation jusqu'à obtention de l'organisation finale. Il communique alors ce tableau à tous pour une validation définitive. Dès qu'une liste est validée de manière définitive, l'organisateur peut passer l'état d'un tableau à "version finale".


Installation
------------

A décrire

Description des données
-----------------------

Lors du démarrage du serveur, la base de données est initialisée et le mot de passe administrateur est généré. Celui-ci est alors affiché dans la console. A chaque redémarrage du serveur, le mot de passe administrateur est affiché dans la console pour rappel.

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
  * etat du tableau (modifications ouvertes à tous, modifications limitées aux administrateurs, version finale)
  * liste des inscrits
  * liste de code de désinscription (1 pour chaque inscrit)
