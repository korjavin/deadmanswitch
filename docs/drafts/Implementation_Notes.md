# Implementation Notes (logical steps)

1. **Secret creation**
   * Owner submits plaintext + N Q&A pairs.
   * Backend:
     1. `DEK = random(32)`
     2. `ciphertext = AES_GCM(DEK, plaintext)`
     3. `shares = shamir_split(DEK, k, N)`
     4. For each answer:
        * `salt = random(16)`
        * `K_i = argon2id(answer, salt, t=4, m=512 MiB)`
        * `C_i = AES_GCM(K_i, share_i)`
     5. Build JSON `{q, salt, C_i}`
     6. `round = drand_now() + X_days`
     7. `questions_blob = tlock_encrypt(JSON, round)`
     8. Store `(ciphertext, questions_blob, k)`

2. **Heartbeat**
   * Endpoint `/alive`:
     * Update `last_seen = now`
     * Recompute `round = now + X_days`
     * Re‑encrypt `questions_blob` for new round (same JSON)

3. **Dispatch**
   * Cron hourly:
     * If `now - last_seen > X_days` and `dispatched = false`:
       * Send signed email with link `/redeem/<note_id>/<token>`
       * Set `dispatched = true`

4. **Redeem page**
   * Fetch `questions_blob`
   * If `tlock_decrypt` returns “too early”, show countdown
   * After unlock, render questions, collect answers
   * Client‑side:
     * For each answer compute `K_i`, decrypt `C_i`
     * When ≥ k shares valid → `DEK = shamir_join`
     * Decrypt ciphertext, display plaintext

5. **Security controls**
   * Argon2id heavy params to slow offline attack
   * Rate‑limit answer trials in browser
   * CSP: no external scripts
   * All sensitive JSON delivered via HTTPS, HSTS

