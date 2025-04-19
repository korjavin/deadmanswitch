# “DeathManSwitch” Service — Questions + Timelock Scheme

## Goal
Post‑mortem delivery of confidential data (crypto seed phrases, bank numbers, mail passwords) to trusted recipients if the owner stops checking‑in.

## Entities
| Item | Description |
|------|-------------|
| **Secret** | File/text encrypted with random `DEK` (AES‑256‑GCM). |
| **Question** | Personal question; only owner & recipient know the answer. |
| **Timelock wrapper** | drand `tlock` encryption of questions; opens after *X* days without heartbeat. |

## Flow
1. Owner creates *N* personal questions (e.g. 10) and enters answers.
2. Generate random `DEK`; encrypt secret.
3. Split `DEK` via Shamir k‑of‑N (k = ceil(⅔ N)).
4. Each share `S_i` encrypted with `K_i = Argon2id(answer_i, salt_i)`.
5. JSON `[ {q,salt,C_i}, … ]` timelock‑encrypted for round `R_0 = now + X days`.
6. On every **alive** ping:
   * compute `R_next = now + X days`;
   * re‑encrypt questions for `R_next`.
7. If owner disappears, round `R_last` reaches the network → questions open.
8. Recipient answers questions; any k correct answers rebuild `DEK`, decrypt secret.
9. If k not reached → secret is lost.

## Trade‑off
A full snapshot lets an attacker brute‑force answers once questions open. Risk accepted because:
* questions are highly personal;
* only questions, not secrets, get revealed;
* skipping external KMS keeps the setup simple and removes a single point of failure.
