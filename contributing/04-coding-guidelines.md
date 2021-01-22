---
title: Coding guidelines
description:
menu:
  docs:
    parent: contributing
    weight: 400
weight: 400
---

## Documenting your code

Please first read the section "Documentation for contributors" in `contributing/00-introduction.md`.

We believe documentation should exist as close to the source code as possible. That makes it more likely to be seen and kept updated.

### Guidelines for commenting

These are roughly our guidelines for where to place your comments:

1. For single pieces of code, the best place to explain it is in comments right next to it
2. For entire files _or large parts of them_, the best place is in comments at the beginning of the file (_not_ in the middle even though they only apply to half the file!)
3. For architecture that extends beyond a single file, as well as useful commands, guidelines, testing notes, etc., create a `README.md` file stored in the given subdirectory
   - Every major directory in the project should have such a file
   - For clarity, they should always start with an `h1` header with the path relative to root (browse the repo for examples)

### Guidelines on things to document

This is a non-exhaustive list of things that are useful to document about your code:

- Include links to any references (such as tutorials, examples, open source code, etc.) that inspired you as your wrote the code
- Explain what other parts of the codebase that the code is used in (or expected to be used in)
- Follow the documentation conventions for Go as that makes the codebase easier to browse with [GoDoc](https://godoc.org/gitlab.com/beneath-hq/beneath) (also applies to other languages)
- Note which functions/structs/variables/constants that were difficult to name and briefly why (that's a great indicator of ambiguity to beware of for people who read the code later)
