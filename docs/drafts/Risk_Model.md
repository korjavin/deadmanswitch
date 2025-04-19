# Risk Model & Motivation

## Motivation
The service is designed to hand over:
* cryptocurrency seed phrases,
* bank / brokerage account numbers,
* primary e‑mail passwords,
* other life‑critical credentials  
to family members if the owner becomes permanently unavailable.

## Assets
| Asset | Value |
|-------|-------|
| Secrets (plaintext) | Financial & personal control |
| Questions / Answers | Secondary; leak aids brute‑force |

## Adversary
* Gains **full disk snapshot** of the server (including DB & code).
* Can keep a clone online or offline.
* Cannot obtain owner’s master password.
* May attempt social engineering on recipients.

## Accepted risks
| Scenario | Impact | Rationale |
|----------|--------|-----------|
| Snapshot + brute‑force succeeds after questions open | Secrets exposed to attacker | Mitigated by high‑entropy personal answers; accepted for simplicity |
| Recipients forget answers | Secret lost | Acceptable; owner chooses k & crafts questions |
| drand compromised (≥ threshold nodes) | Early reveal | Unlikely, monitored by community |
| Email phishing on delivery | Recipient misled | Use DKIM + signed links; educate recipients |

## Non‑accepted risks
* Compromise before questions open still protected (attacker doesn’t know answers).
* DoS on email sending considered non‑critical (recipients can visit site manually).

