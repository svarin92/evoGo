# evoGo : An Evolutionary Symbolic Engine Based on Grammatical Evolution (GE)

---

## **Description**

**evoGo** is a proof-of-concept **evolutionary symbolic engine** developed in **Go**, inspired by the principles of **Grammatical Evolution (GE)**. It is designed to generate and optimize **symbolic structures** (such as expressions or programs) based on **formal grammars** (specifically, Context-Free Grammars or CFGs). Unlike traditional genetic algorithms (GAs), which manipulate numerical values, **evoGo** employs a symbolic approach to produce structured and interpretable solutions.

Although the project shares many features with a standard grammatical evolution engine — notably the use of a symbolic search space, an evolutionary cycle (selection, crossover, mutation), and a separation between genotype and phenotype — it is part of a broader vision. Grammatical evolution relies strictly on sequential choices made via rewriting rules (typically written in BNF — Backus-Naur form), whereas a symbolic evolutionary engine like evoGo can incorporate more flexible correction and manipulation mechanisms that are semantically driven or implemented through dedicated filtering and correction phases. Thus, grammatical evolution can be viewed as a formal instantiation of a wider symbolic paradigm that evoGo seeks to explore and implement in Go.

### **Key Features**

- **Symbolic approach**: generation of solutions as structures derived from formal grammars.
- **Evolutionary optimization**: use of a genetic algorithm loop (selection, crossover, mutation) to optimize phenotypes.
- **Immune response**: integration of an immune system comprising three correction mechanisms (non-coding RNA, lymphocytes, repair enzymes) to ensure solution validity and robustness.
- **Performance**: implementation in **Go** for fast and efficient execution.

---

## **Scientific Context**

**evoGo** draws inspiration from foundational work in **Grammatical Evolution (GE)**:

- **PonyGE**: a Grammatical Evolution framework in Python, widely used in the literature for its ability to generate evolvable programs.
- **Michael O'Neill** and Conor Ryan: pioneers of GE and authors of the book *Grammatical Evolution: Evolutionary Automatic Programming in an Arbitrary Language* (Kluwer Academic Publishers, 2003), which establishes the theoretical foundations of this approach.

---

## **Architecture**

### **The grammatical evolution loop**

The core of evoGo relies on a modular architecture and a continuous iterative loop, ensuring smooth transitions between the genotype (symbolic expressions in BNF format) and the phenotype (the final system, whether a graphical interface or a network of agents):

1. **Generation**: creation of an initial population of **genomes** (codon sequences).
2. **Derivation**: transformation of genomes into **phenotypes** (symbolic expressions) via a formal grammar.
3. **Evaluation**: calculation of **fitness** for each phenotype.
4. **Selection**: selection of the best-performing individuals for reproduction.
5. **Genetic operators**: application of **crossover** and **mutation** to generate a new population.
6. **Immune response**: semantic validation and automatic correction of invalid individuals to ensure phenotypes remain viable and executable.

### **The Immune System**

**evoGo** incorporates a novel **embryonic immune system** composed of three biologically inspired genomic correction mechanisms.

#### **Non-coding RNAs (`CorrectByTemplate`)**

This corrector acts as a **dynamic regulator** of the genome. It adjusts phenotypes using templates (e.g., truncating invalid sequences) and ensures their compliance with grammar rules.

**Log example**:

```
Starting decoding for block: [0 0 -1 1 0 1 1 0 1 0 1 6 5 1 7 61 56 38 0 88 99 96 55 118 64 107 103 9 119 45 60 35 55 90 39 0 70 46 114 42 51 63 64 45 8 93 41 66 67 88 74 1 55 39 106 74 61 48 93 114 105 64 86 120 24 31 69 28 32 46 13 46 111 29 93 18 72 56 72 9 42 101 5 57 6 47 21 110 116 25 121 108 97 24 112 38 27 46 44 55 56 57 74 81 81 29 14 16 123 39 125 49 85 94 69 123 38 27 89 2 22 84 16 116 113 30 123], startRule: "grammar"
...
Using non-coding RNA letter_letter_letter_letter_letter_letter_exp: [letter 1 letter 1 letter 1 letter 1 letter 1 letter 1]
...
Final productions: [[string 1] [letter 1 string_tail 1] [letter 1 letter 1 letter 1 letter 1 letter 1 letter 1] [consonant 1] [vowel 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1] [g 0] [e 0] [l 0] [d 0] [e 0] [n 0]]
```

#### **Lymphocytes (`CorrectByGenome`)**

This mechanism performs a global correction of individuals by identifying and repairing invalid genomic sequences, much like the elimination of defective cells by T-lymphocytes.

**Log example**:

```
Starting - usePhenotype=false, individual.phenotype=glllon
Using implicit production history: [[string 1] [letter 1 string_tail 1] [consonant 1] [consonant 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1] [g 0] [l 0] [l 0] [l 0] [o 0] [n 0]]
Found factorizable sequence at index 2 to 7 (length 6): [[consonant 1] [consonant 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1]]. Matched rule: letter
Built expansion for rule "letter" (length 6): [letter 1 letter 1 letter 1 letter 1 letter 1 letter 1]
Inserted expansion at index 2 for sequence [[consonant 1] [consonant 1] [consonant 1] [consonant 1] [vowel 1] [consonant 1]]
Encoded genome=[0 0 -1 1 1 1 1 0 1 0 6 6 6 2 7 ...]
```

#### **Repair enzymes (`CorrectByGrammaticalPaths`)**

This mechanism operates in a **local and targeted** manner to repair invalid structural patterns (e.g., malformed grammatical paths), much like DNA repair enzymes.

**Log example**:

```
Starting decoding for block: [1 0 1], startRule: "grammar"
Codon 0 decoded to production by codon index 0: [string 1] for symbol "grammar"
Codon 0 decoded to production by codon index 1: [letter 1 string_tail 1] for symbol "string"
Codon 1 decoded to production by codon index 2: [consonant 1] for symbol "letter"
Final productions: [[string 1] [letter 1 string_tail 1] [consonant 1]]
```

---

## **Project Structure**

```bash
evoGo/
├── config/                  # Project configuration
├── controller/genomizer.go  # Grammatical Evolution (GE) engine
├── controller/immunizer.go  # Immune system (genomic corrector)
├── controller/serializer.go # Serialization of grammars and genomes
├── evaluator/               # Phenotype evaluation
├── examples/                # Usage examples (e.g., EBNF grammars)
├── ge/                      # Evolution cycle
├── grammar                  # Grammar specialization
├── interfaces/              # Interfaces and contracts (e.g., IGrammar, IIndividual)
├── model/                   # Grammar models (e.g., RuleModel, SequenceModel)
├── operators/               # Genetic operators (crossover, mutation)
├── patterns/                # Design patterns (Visitor, Builder, Notifier)
├── prolog/                  # Hybridization with Prolog for symbolic reasoning
├── renderer/                # Rendering and visualization engine
├── utils/                   # Utilities
├── LISEZ_MOI.md             # Main documentation in French
└── README.md                # Main documentation
```

---

## **Execution Example**

### **Evolution towards the "golden" phenotype**

The following example illustrates a **complete execution** of the **evoGo** evolutionary loop, using a simple grammar designed to generate the word **"golden"**. The goal is to demonstrate how a **valid phenotype** emerges from a random initial population and how its **underlying structure** reveals a **semantic rule**.

**Log excerpt**:

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

### **What do these results tell us?**

1. **Phenotype structure**: The genome encodes a sequence of codons producing the word "golden." The phenotype is generated with a **fitness of 1.00**, meaning it perfectly matches the target. The **grammatical path** shows how each letter (`g`, `o`, `l`, `d`, `e`, `n`) is derived via production rules (e.g., `consonant → g`, `vowel → o`).
2. **Symbolic proof and semantic rule**: The grammatical path reveals a **recurring structure**: an alternation between consonants and vowels. This structure can be generalized into a **semantic rule**: *"To form a valid word like 'golden,' a 6-letter sequence is required with a consonant-vowel-consonant-consonant-vowel-consonant alternation."*
3. **Transition to symbolic evaluation**: **Fitness** becomes a **logical proof**: if the phenotype is **"golden,"** its grammatical structure proves that it adheres to the rules of the grammar.
4. **Learning curve**: Each generation learns to generate structures closer to the target, thanks to genetic operators and immune corrections.

---

## **Towards a Self-Organizing and Self-Learning System**

The prospects for evoGo center on a profound convergence between distributed self-organization and formal symbolic reasoning, laying the groundwork for a truly autonomous system. Moving beyond the strict confines of traditional genetic algorithm optimization — typically used to describe the generative structures of grammatical evolution — the architecture aims to merge the adaptive plasticity of cellular populations with the inferential rigor of symbolic logic, thereby unifying behavioral emergence and semantic correctness within a single Go-based ecosystem.

### **Genetic Homeostasis and Cellular Dynamics**

To enhance the system's structural robustness, the primary approach involves hybridizing evoGo with the [go-actor](https://github.com/vladopajic/go-actor) actor framework. By encapsulating genetic code and phenotypes within a cellular structure (DigitalCell), the population transforms into a network of autonomous, communicating agents. In this paradigm, the classic phases of selection and mutation are no longer driven statically or centrally but emerge dynamically from local interactions. Each cell manages its own homeostasis by responding to its environment through feedback loops inspired by cellular automata. This distributed immune mechanism acts as a continuous corrective filter, ensuring the viability and coherence of genetic expressions prior to global validation, thereby endowing the system with adaptive plasticity in the face of environmental disturbances.

### **Advanced Symbolic Phenotype Evaluation and Logical Inference**

Alongside self-organization, phenotypic validation can be enhanced with high-level formal semantics through the integration of a [Prolog](https://github.com/ichiban/prolog) engine — or expert system — embedded in Go. Unlike numerical or syntactic fitness functions, this symbolic evaluation analyzes generated structures using knowledge bases and explicit logical rules. The system goes beyond merely measuring quantitative performance; it assesses structural validity, contextual coherence, and adherence to complex constraints — such as geometric or aesthetic criteria. This combination of evoGo’s evolutionary exploration and Prolog’s deductive power enables the engine to guide evolution not through simple statistical approximation, but through structured symbolic reasoning, paving the way for genuine cognitive autonomy.

These possibilities and this exploratory power are made possible precisely because evoGo breaks free from a purely grammatical framework. It establishes itself as a **true evolutionary symbolic engine**, capable not only of manipulating formal structures but also of shaping their semantics, correctness, and emergent behavior.

---

## **Use Cases**

- **Program generation**: creation of source code based on language grammars (Python, Go).
- **Symbolic optimization**: solving complex problems requiring mathematical or logical expressions.
- **Academic research**: study of symbolic evolution mechanisms and self-learning reasoning.

---

## **References**

1. [PonyGE](https://github.com/jmmcd/ponyge): Grammatical Evolution in Python.
2. Michael O'Neill and Conor Ryan, [Grammatical Evolution: Evolutionary
Automatic Programming in an Arbitrary Language](https://link.springer.com/book/10.1007/978-1-4615-0447-4), Kluwer Academic
Publishers, 2003.
3. [Ichiban/Prolog](https://github.com/ichiban/prolog): The only reasonable scripting engine for Go.
4. [go-actor](https://github.com/vladopajic/go-actor): A lightweight library for writing concurrent programs in Go using the Actor model.
5. **Go Libraries** :
  - [kong](https://github.com/alecthomas/kong): A command-line parser for Go.
  - [participle/v2/ebnf](https://github.com/alecthomas/participle) : A parser library for Go.

---

## **Contribution**

Contributions are welcome via issues or pull requests.

---

## **License**

This project is distributed under the **MIT License**. See the [LICENSE](LICENSE) file for details.

---

## **Contact**

For any questions or collaboration:

- **Stéphane Varin** : [svarin92@gmail.com](mailto:svarin92@gmail.com)
- **GitHub Repository** : [github.com/svarin92/evoGo](https://github.com/svarin92/evoGo)