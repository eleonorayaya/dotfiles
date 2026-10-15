---
name: writing-style
description: Use when writing documents, copy, pull request body etc
---
# Writing Structure

Style is voice: tone, phrasing, punctuation. This is structure: what goes where, how much of it, and in what form.

Applies to anything another person will read. Two tiers:

- **Always** — Slack, Jira, PR bodies, commit messages, docs, specs, decks
- **Long-form** — docs, specs, decks only. A Slack message has no appendix; skip those rules for short-form.

The default to correct is exhaustiveness. Detail-completeness reads as diligence to the author and as unreviewable to the reader. Bias short. That is the correct direction to err.

---

## 1. Establish the audience (always)

State the read in one line before writing. Don't interrogate:

> Writing for: Marco, deciding whether to fund the spike, ~10 minutes, reads it alone.

| Variable | What it determines |
|---|---|
| Who reads it, and what they already know | What can be assumed |
| What they decide as a result | What's in scope |
| Reading occasion — walked live / skimmed pre-meeting / read alone at depth | The artifact set |
| How much time they'll give it | The hardest constraint, and the one most often skipped |

Ask an actual question only when the answer would materially change the artifact and context genuinely doesn't supply it. That's rarer than it feels.

**Default when it can't be inferred:** peer-level engineer, ten minutes, one decision to make.

**Persist it.** A revision inherits the established audience. Don't re-open the altitude question mid-artifact and don't drift back up. State the audience read at the top of the working record, not in the reader artifact.

## 2. Choose the artifact set (long-form)

Detail is a placement decision, not a volume decision. Each layer answers a reading occasion, and content moves *down* until it reaches the layer whose occasion needs it. The spine is the last resort, not the default.

| Layer | Reading occasion | Contents |
|---|---|---|
| Spine — deck, spec body | Walked live, or skimmed in ten minutes | The argument and the ask |
| Appendix | Pulled when a specific person challenges a specific thing | Rejected alternatives, contested calls, the table someone will want |
| Supporting doc | Read alone by whoever implements or approves | Full mechanism |
| Working record | Retrieved later by you or an agent | Provenance, dead ends, everything considered |

**Split off a working record when** the work involved exploration whose conclusions aren't self-evident from the output, or a real alternative was considered and rejected. Otherwise one doc.

The working record is optimized for retrieval, not for reading front to back: dense, evidence-anchored, things named with IDs and dates. Its highest-value content is the rejected alternatives and *why*, which is exactly what a polish pass deletes. It is **not a mirror** — the reader artifact is canonical for the decision, and the working record is append-mostly, allowed to go stale on anything but its own content.

**Appendix discipline:** one item per named anticipated challenge, and say it isn't walked. An appendix that receives everything cut from the spine is a landfill nobody can pull the right thread from live.

**One layer is canonical for each fact.** Others reference it rather than restate it. Restating is how counts and definitions drift between layers.

## 3. Draft at altitude (long-form)

- **For each piece of detail, name the decision it changes for this reader.** If there isn't one, it moves down a layer.
- **Progressive disclosure.** A reader who stops at any point should have a coherent understanding at that altitude, not a truncated one. Each layer completes before the next descends.
- **Why before how, enforced by ordering.** A section explaining a mechanism before the reader knows what decision it serves is in the wrong place. Move it; don't bolt on a why paragraph.
- **Headings carry the argument.** Topic labels (Architecture, Data Model, Migration) mean there's no narrative, just a filing system. Reading the headings alone should give the argument.
- **Slides are not documents.** A slide body is a claim plus its evidence. The script goes in speaker notes, not on the slide.
- **Tech specs are not exempt.** A spec's job is to let a reviewer decide, not to preempt every question they might have. A spec too long to read doesn't get the scrutiny it exists to attract.

## 4. Choose the form (always)

Route by the reader's question, not by what's cheapest to generate.

| Reader's question | Form |
|---|---|
| How do these things relate? | Diagram |
| How do these options compare? | Table |
| What do I write, and what comes back? | Artifact — YAML, API request + response, CLI invocation + output, config |
| What changes for me? | Before/after artifact pair |
| Why this and not that? | Prose |

Triggers:

- A relationship among 3+ things is a table or diagram, not a paragraph.
- Comparing 2+ options across 2+ dimensions is a table. Comparative prose makes the reader hold both columns in their head.
- A flow with 3+ steps, or any branching, is a diagram or numbered steps.
- Explaining an abstraction leads with one concrete named instance, then generalizes. One worked example beats a complete description.
- Examples use real names from the actual system. Never `Service A`, `foo`, `my-app`.

Artifact rules:

- **Real and plausible.** Real field names, real service names, values that would actually work. `foo: bar` teaches nothing and signals the author didn't check.
- **Minimal, with elisions marked** (`# ...`). Trim to the fields that carry the point.
- **Never invent an artifact for a system you haven't verified.** Read the schema or the API first, or label it illustrative and say so. Confidently wrong YAML is worse than verbose prose.
- **Synthetic values only.** No real credentials, account IDs, customer or employee names, or other PII in an example.

Medium routing: HTML → inline SVG · markdown → mermaid or table · Google Slides → shapes · Slack → aligned short lines or a code block, since its table rendering is poor · Jira → its own table markup.

**Escape hatch:** prose carries argument and causation, and a table can't express "because." Don't tabulate a claim, and don't build a table whose cells are sentences — that's a list in costume.

## 5. Reveal in the reader's order (always)

The reader has none of the author's context, and the order things were discovered in is close to the worst order to present them in.

- **Takeaway first, mechanism last.** Claim → why it matters → how it works → details. The conclusion goes up front so the reader knows where they're heading; the mechanism waits until they care.
- **No forward references.** A section leaning on a concept defined later means one of the two is misplaced.
- **Define at first use**, not in a terminology block up front.
- **One new concept at a time.** Two or three arriving together means an incremental build got collapsed.
- **A term means one thing throughout.** Same word, same referent, every section.

## 6. PR descriptions (always)

The body frames the diff, it does not restate it. The reviewer reads the files. The description tells them what reading the files won't.

**Inclusion test:** would someone who read every changed file still not know this? If they would know it, cut it.

| Cut | Keep |
|---|---|
| The file list, per-file summaries | Config or state outside the repo that the change depends on |
| Anything already in a code comment | Ordering and placement decisions whose reason isn't visible locally |
| Verification matrices, probe logs, environment archaeology | Behaviour changes to other developers' local workflow |
| Explanations of standard tooling | Deliberate omissions, marked as decisions so they don't read as oversights |

**Assume colleagues' knowledge.** No explaining Buildkite, GitHub status checks, Helm, `go test`. Teaching a reader their own CI is the worst version of over-explaining, and reads as condescension rather than thoroughness.

**Lead with the implication, not the identifier.** "Acceptance tests are skipped in CI," not "`TestAccLifecycle` is now guarded." A symbol name makes the reader look something up before they know whether they care. Exception: when the identifiers *are* the contrast (`go test` vs `make test`).

**One claim per bullet.** Any 4+ sentence block of reasoning is a fused argument, so split it. Bolded leads run four to six words and stay roughly even in length across the list, because they are scan targets, not sentences in bold.

**Restructure before trimming.** "Too long" is usually shape, not word count. Cutting words while keeping dense paragraphs does not fix it, and burns a review round.

**Compression exposes shaky claims.** A claim that reads oddly once compressed to one line is often wrong rather than badly worded. Check the fact before rewording it.

**Use the phrasing the author suggested,** not a paraphrase. Paraphrase drifts back from consequence ("acceptance tests are skipped in CI") toward mechanism ("acceptance tests now opt-in").

Evidence of testing is one or two lines: what was run, what it proved. Anything longer is working record.

## 7. Check before handing it over (always)

**Run this list before showing a draft, not after a reviewer reacts to one.** These rules only bind if the check happens between drafting and handing over, so never present a first draft. Cut first, then show. A second draft that arrives only because a reviewer pushed back means the check was skipped.

- What was cut? If nothing was, it wasn't edited.
- Does the first sentence carry new information, or restate the question?
- Is there a closing summary that repeats what was just read? Delete it.
- Do counts, names, and definitions reconcile across every section and layer?
- Is the takeaway landed, or is the recommendation withheld? (On a slide: is there exactly one?)
- Did anything get bullets it didn't need because its neighbors had three? Real content has uneven weight.
- Is nesting past two levels doing work, or standing in for thought?
- Is so much bold that nothing is emphasized?
- If this is a revision: is it the same length or shorter?
