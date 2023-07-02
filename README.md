# Context Free Grammar (CFG)

## Formal Definition

A context-free grammar `G` is a 4-tuple `G = (V, Σ, R, S)` where:

1. `V` is a finite set; each element `v ∈ V` is called a variable. Each variable represents a different type of phrase
   or clause in the sentence.
2. `Σ` is a finite set of terminals, disjoint from `V`, which make up the actual content of the sentence. The set of
   terminals is the alphabet of the language defined by the grammar `G`.
3. `R` is a finite relation in `(V × (V ∪ Σ)*)`. The members of `R` are called productions of the grammar (symbolized
   by `P`).
4. `S` is the start variable, used to represent the whole sentence. It must be an element of `V`.

## References

- [Wikipedia](https://en.wikipedia.org/wiki/Context-free_grammar)