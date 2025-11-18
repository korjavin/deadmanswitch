# TASK-009: Remove drand/Timelock Mechanism 

## Priority: MEDIUM 🟡

## Status: Not Started

## Category: Feature Removal / Code Cleanup

## Description

Remove the entire drand-based timelock encryption and secret questions feature from the codebase. This includes the cryptographic implementation, database layer, web handlers, UI, scheduler tasks, and all related code (~2,500 lines total).

## Problem

**Current State:**
The project includes a sophisticated but unwanted feature for secret questions using:
- **Timelock encryption** - Secrets locked until a specific time (using drand rounds)

**Why Remove It:**
User has lost interest in this feature due to:
- Complexity of implementation and maintenance
- Additional dependency on drand network concepts
- Over-engineering for the core use case

## Scope of Removal

### 1. Cryptography Layer (~600 lines)


**Code Removed:**
- `TimelockEncrypt()` - Encrypt with drand round
- `TimelockDecrypt()` - Decrypt after round available
- `CalculateDrandRound()` - Calculate drand round for timestamp
