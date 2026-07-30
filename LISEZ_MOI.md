# evoGo : Moteur Symbolique Évolutionnaire Basé sur l'Évolution Grammaticale (GE)

---

## **Description**

**evoGo** est une preuve de concept d'un **moteur symbolique évolutionnaire** développé en **Go**, inspiré des principes de l’**Évolution Grammaticale (GE)**. Il est conçu pour générer et optimiser des **structures symboliques** (expressions, programmes) à partir de **grammaires formelles** (CFG - Context-Free Grammar). Contrairement aux algorithmes génétiques (AG) traditionnels, qui manipulent des valeurs numériques, **evoGo** utilise une approche symbolique pour produire des solutions structurées et interprétables. 

Bien que le projet partage de nombreux points communs avec un moteur d'évolution grammaticale classique — notamment l'utilisation d'un espace de recherche symbolique, d'un cycle évolutif (sélection, croisement, mutation) et d'une séparation entre génotype et phénotype —, il s'inscrit dans une vision plus large. L'Évolution Grammaticale repose purement sur des choix séquentiels via des règles de réécriture (généralement rédigées au format BNF - Backus-Naur form), tandis qu'un moteur symbolique évolutionnaire comme evoGo peut intégrer des mécanismes de correction et de manipulation plus flexibles, orientés sémantiquement ou par des phases de filtrage/correction dédiées. Ainsi, l'évolution grammaticale peut être considérée comme une instanciation formelle d'un paradigme symbolique plus vaste qu'evoGo cherche à explorer et à modeler en Go.

### **Caractéristiques principales**

- **Approche symbolique** : génération de solutions sous forme de structures dérivées de grammaires formelles.
- **Optimisation évolutionnaire** : utilisation d'une boucle d’algorithme génétique (sélection, croisement, mutation) pour optimiser les phénotypes.
- **Réponse immunitaire** : intégration d’un système immunitaire composé de trois mécanismes de correction (ARN non codants, lymphocytes, enzymes réparateurs) pour garantir la validité et la robustesse des solutions.
- **Performance** : implémentation en **Go** pour une exécution rapide et efficace.

---

## **Contexte scientifique**

**evoGo** s’inspire des travaux fondateurs en **Évolution Grammaticale (GE)** :

- **PonyGE** : framework d’Évolution Grammaticale en Python, largement utilisé dans la littérature pour ses capacités à générer des programmes évolutifs.
- **Michael O'Neill** et Conor Ryan : pionniers de la GE, auteurs de l’ouvrage *Grammatical Evolution*: Evolutionary Automatic Programming in an Arbitrary Language (Kluwer Academic Publishers, 2003), qui pose les bases théoriques de cette approche.

---

## **Architecture**

### **La boucle d'évolution grammaticale**

Le cœur d’evoGo repose sur une architecture modulaire et une boucle itérative continue, assurant la fluidité des transitions entre le génotype (les expressions symboliques sous forme de BNF) et le phénotype (le système final, qu'il s'agisse d'une interface graphique ou d'un réseau d'agents) :

1. **Génération** : création d’une population initiale de **génomes** (séquences de codons).
2. **Dérivation** : transformation des génomes en **phénotypes** (expressions symboliques) via une grammaire formelle.
3. **Évaluation** : calcul du **fitness** pour chaque phénotype.
4. **Sélection** : sélection des individus les plus performants pour la reproduction.
5. **Opérateurs génétiques** : application de **croisements** et **mutations** pour générer une nouvelle population.
6. **Réponse immunitaire** : validation sémantique et correction automatique des individus invalides pour garantir des phénotypes toujours viables et exécutables.

### **Le  système immunitaire**

**evoGo** intègre un **système immunitaire embryonnaire** original composé de trois mécanismes de correction génomique, inspirés de la biologgie.

#### **ARN non codants (`CorrectByTemplate`)**

Ce correcteur agit comme un **régulateur dynamique** du génome. Il ajuste les phénotypes via des *templates* (ex: tronquage de séquences invalides) et garantit leur conformité aux règles de la grammaire.

**Exemple de log** :

```
Starting decoding for block: [0 0 -1 1 0 1 1 0 1 0 1 6 5 1 7 61 56 38 0 88 99 96 55 118 64 107 103 9 119 45 60 35 55 90 39 0 70 46 114 42 51 63 64 45 8 93 41 66 67 88 74 1 55 39 106 74 61 48 93 114 105 64 86 120 24 31 69 28 32 46 13 46 111 29 93 18 72 56 72 9 42 101 5 57 6 47 21 110 116 25 121 108 97 24 112 38 27 46 44 55 56 57 74 81 81 29 14 16 123 39 125 49 85 94 69 123 38 27 89 2 22 84 16 116 113 30 123], startRule: "grammar"
...
Using non-coding RNA letter_letter_letter_letter_letter_letter_exp: [letter 1 letter 1 letter 1 letter 1 letter 1 letter 1]
...
Final productions: [[string 1] [letter 1 string_tail 1] [letter 1 letter 1 letter 1 letter 1 letter 1 letter 1] [consonant 1] [vowel 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1] [g 0] [e 0] [l 0] [d 0] [e 0] [n 0]]
```

#### **Lymphocytes (`CorrectByGenome`)**

Ce mécanisme effectue une correction globale des individus en identifiant et réparant les séquences génomiques invalides, à l'image de l’élimination des cellules défectueuses par les lymphocytes T.

**Exemple de log** :

```
Starting - usePhenotype=false, individual.phenotype=glllon
Using implicit production history: [[string 1] [letter 1 string_tail 1] [consonant 1] [consonant 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1] [g 0] [l 0] [l 0] [l 0] [o 0] [n 0]]
Found factorizable sequence at index 2 to 7 (length 6): [[consonant 1] [consonant 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1]]. Matched rule: letter
Built expansion for rule "letter" (length 6): [letter 1 letter 1 letter 1 letter 1 letter 1 letter 1]
Inserted expansion at index 2 for sequence [[consonant 1] [consonant 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1]]
Encoded genome=[0 0 -1 1 1 1 1 0 1 0 6 6 6 2 7 ...]
```

#### **Enzymes réparateurs (`CorrectByGrammaticalPaths`)**

Ce mécanisme intervient de manière **locale et ciblée** pour réparer des motifs structurels invalides (ex: chemins grammaticaux mal formés), à l'image des enzymes de réparation de l'ADN.

**Exemple de log** :

```
Starting decoding for block: [1 0 1], startRule: "grammar"
Codon 0 decoded to production by codon index 0: [string 1] for symbol "grammar"
Codon 0 decoded to production by codon index 1: [letter 1 string_tail 1] for symbol "string"
Codon 1 decoded to production by codon index 2: [consonant 1] for symbol "letter"
Final productions: [[string 1] [letter 1 string_tail 1] [consonant 1]]
```

---

## **Structure du Projet**

```bash
evoGo/
├── config/                  # Configuration du projet
├── controller/genomizer.go  # Moteur d'évolution grammaticale (GE)
├── controller/immunizer.go  # Système immunitaire (correcteur génomique)
├── controller/serializer.go # Sérialisation des grammaires et génomes
├── evaluator/               # Évaluation des phénotypes
├── examples/                # Exemples d'utilisation (ex: grammaires EBNF)
├── ge/                      # Cycle d'évolution
├── grammar                  # Spécialisation des grammaires
├── interfaces/              # Interfaces et contrats (ex: IGrammar, IIndividual)
├── model/                   # Modèles de grammaires (ex: RuleModel, SequenceModel)
├── operators/               # Opérateurs génétiques (croisement, mutation)
├── patterns/                # Patterns de conception (Visitor, Builder, Notifier)
├── prolog/                  # Hybridation avec Prolog pour le raisonnement symbolique
├── renderer/                # Moteur de rendu et de visualisation
├── utils/                   # Utilitaires
├── LISEZ_MOI.md             # Documentation principale en français
└── README.md                # Documentation principale
```

---

## **Exemple d'exécution**

### **Évolution vers le phénotype "golden"**

L'exemple suivant illustre une **exécution complète** de la boucle évolutive d’**evoGo**, avec une grammaire simple visant à générer le mot **"golden"**. L'objectif est de montrer comment un **phénotype valide** émerge d'une population initiale aléatoire et comment sa **structure sous-jacente** révèle une **règle sémantique**.

**Extrait de log** :

```plaintext
--- Génération 0 ---
Population initiale : [glllon, glden, golxen, gooden, gloden, ...]
Fitness moyen : 0.12
Meilleur phénotype : "glllon" (fitness = 0.45)

--- Génération 10 ---
Population : [gollen, golden, golxen, gooden, gloden, ...]
Fitness moyen : 0.45
Meilleur phénotype : "gollen" (fitness = 0.80)

--- Génération 25 ---
Population : [golden, golden, golxen, golden, gloden, ...]
Fitness moyen : 0.88
Meilleur phénotype : "golden" (fitness = 1.0)

=== Best Individual (Generation 25) ===
Phenotype: golden
Fitness: 1.00
Genome: [0 0 -1 1 0 1 1 0 1 0 2 6 5 1 7 51 112 36 105 55 74 69 115 69 59 120 48 103 30 112 54 101 46 91 7 23 103 105 107 12 84 58 12 11 56 4 7 67 67 10 38 71 94 49 59 4 39 91 111 25 39 123 65 21 39 86 87 6 17 41 92 28 16 52 50 126 48 18 89 45 114 16 21 51 57 120 62 86 46 48 98 74 102 110 48 11 54 122 115 54 42 110 5 36 36 98 24 36 44 85 104 105 28 28 97 51 1 6 17 100 75 6 26 21 34 14 112], Length: 127

--- Grammatical Path of the Phenotype ---
Phenotype: golden (Fitness: 1.00)
PrintGrammaticalDerivation: history [[string 1] [letter 1 string_tail 1] [consonant 1] [vowel 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1] [g 0] [o 0] [l 0] [d 0] [e 0] [n 0]]
 1. string (fitness: 0.77) *
 2. letter string_tail (fitness: 0.77) *
 3. consonant (fitness: 0.77) *
 4. vowel (fitness: 0.77) *
 5. consonant (fitness: 0.77) *
 6. consonant (fitness: 0.77) *
 7. vowel (fitness: 0.77) *
 8. consonant (fitness: 0.77) *
 9. 'g' (fitness: 0.77)
10. 'o' (fitness: 0.00)
11. 'l' (fitness: 0.77)
12. 'd' (fitness: 0.69)
13. 'e' (fitness: 0.77) *
14. 'n' (fitness: 0.43)
----------------------------------
2026/07/27 09:56:16 Successfully exported derivation to best_individual.dot
Best individual: phenotype = golden, fitness = 1.00, Used Codons: 14
```

### **Que nous disent ces résultats ?**

1. **Structure du phénotype** : le génome encode une séquence de codons produisant le mot "golden". Le phénotype est généré avec un **fitness de 1.00**, ce qui signifie qu'il correspond parfaitement à l'objectif. Le **chemin grammatical** montre comment chaque lettre (`g`, `o`, `l`, `d`, `e`, `n`) est dérivée via des règles de production (ex: `consonant → g`, `vowel → o`).
2. **Preuve symbolique et règle sémantique** : le chemin grammatical révèle une **structure récurrente** : une alternance entre consonnes et voyelles. Cette structure peut être généralisée en une **règle sémantique** : *"Pour former un mot valide comme 'golden', il faut une séquence de 6 lettres avec une alternance consonant-voyelle-consonant-consonant-voyelle-consonant."*
3. **Transition vers une évaluation symbolique** : le **fitness** devient une **preuve logique** : si le phénotype est **"golden"**, sa structure grammaticale prouve qu'il respecte les règles de la grammaire.
4. **Courbe d'apprentissage** : chaque génération apprend à générer des structures plus proches de l'objectif, grâce aux opérateurs génétiques et aux corrections immunitaires.

---

## **Vers un système auto-organisé et auto-apprenant**

Les perspectives d'evoGo s'articulent autour d'une convergence profonde entre l'auto-organisation distribuée et le raisonnement symbolique formel, posant les bases d'un système véritablement autonome. En dépassant le cadre strict de l'optimisation par algorithme génétique traditionnel, utilisé pour décrire les structures de production de l'évolution grammaticale, l'architecture ambitionne de fusionner la plasticité adaptative des populations cellulaires et la rigueur inférentielle de la logique symbolique, unifiant ainsi au sein d'un même écosystème en Go l'émergence comportementale et la correction sémantique.

### **Homéostasie génétique et dynamiques cellulaires**

Pour renforcer la robustesse structurelle du sytème, la première orientation réside dans l'hybridation d'evoGo avec le framework d'acteurs [go-actor](https://github.com/vladopajic/go-actor). En encapsulant le code génétique et les phénotypes au sein d'une structure cellulaire (DigitalCell), la population se transforme en un réseau d'agents autonomes et communicants. Dans ce paradigme, les phases classiques de sélection et de mutation ne sont plus pilotées de manière statique ou centralisée, mais émergent dynamiquement d'interactions locales. Chaque cellule gère sa propre homéostasie en réagissant à son environnement grâce à des boucles de rétroaction inspirées des automates cellulaires. Ce mécanisme immunitaire distribué agit comme un filtre correcteur permanent, garantissant la viabilité et la cohérence des expressions génétiques avant même leur validation globale, ce qui confére au système une plasticité adaptative face aux perturbations de l'environnement.

### **Évaluation symbolique avancée des phénotypes et inférence logique**

En parallèle de l'auto-organisation, la validation phénotypique peut être augmentée d'une sémantique formelle de haut niveau grâce à l'intégration d'un moteur [Prolog](https://github.com/ichiban/prolog) ou système expert, embarqué en Go. Contrairement aux fonctions de fitness numériques ou syntaxiques, cette évaluation symbolique analyse les structures générées à l'aide de bases de connaissances et de règles logiques explicites. Le système ne se contente plus de mesurer une performance quantitative, il interroge la validité structurelle, la cohérence contextuelle et le respect de contraintes complexes — telles que des critères géométriques ou esthétiques. Cette combinaison entre l'exploration évolutionnaire d'evoGo et la puissance déductive de Prolog permet au moteur de guider l'évolution non plus par simple approximation statistique, mais par un raisonnement symbolique structuré, ouvrant la voie à une véritable autonomie cognitive.

Ces perspectives et cette puissance d'exploration sont rendues possibles précisément parce que evoGo s'affranchit du cadre purement grammatical. Il s'affirme comme un **véritable moteur symbolique évolutionnaire**, capable non seulement de manipuler des structures formelles, mais encore d'en façonner la sémantique, la correction et l'émergence comportementale.

---

## **Cas d'Usage**

- **Génération de programmes** : création de code source à partir de grammaires de langages (Python, Go).
- **Optimisation symbolique** : résolution de problèmes complexes nécessitant des expressions mathématiques ou logiques.
- **Recherche académique** : étude des mécanismes d’évolution symbolique et de raisonnement auto-apprenant.

---

## **Références**

1. [PonyGE](https://github.com/jmmcd/ponyge) : Grammatical Evolution in Python.
2. Michael O'Neill and Conor Ryan, [Grammatical Evolution: Evolutionary
Automatic Programming in an Arbitrary Language](https://link.springer.com/book/10.1007/978-1-4615-0447-4), Kluwer Academic
Publishers, 2003.
3. [Ichiban/Prolog](https://github.com/ichiban/prolog) : The only reasonable scripting engine for Go.
4. [go-actor](https://github.com/vladopajic/go-actor) : A lightweight library for writing concurrent programs in Go using the Actor model.
5. **Bibliothèques Go** :
  - [kong](https://github.com/alecthomas/kong) : A command-line parser for Go.
  - [participle/v2/ebnf](https://github.com/alecthomas/participle) : A parser library for Go.

---

## **Contribution**

Les contributions sont bienvenues via des issues ou des Pull Requests.

---

## **Licence**

Ce projet est distribué sous la **licence MIT**. Consultez le fichier [LICENSE](LICENSE) pour plus de détails.

---

## **Contact**

Pour toute question ou collaboration :

- **Stéphane Varin** : [svarin92@gmail.com](mailto:svarin92@gmail.com)
- **Dépôt GitHub** : [github.com/svarin92/evoGo](https://github.com/svarin92/evoGo)