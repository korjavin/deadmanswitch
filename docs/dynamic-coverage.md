# Dynamic Code Coverage Requirements

This document explains our approach to code coverage requirements that evolve over time.

## Philosophy

We believe that test coverage is important for ensuring code quality and preventing regressions. However, we also recognize that:

1. Achieving high test coverage takes time
2. Setting an immediate high threshold can block progress
3. Different parts of the codebase may require different levels of coverage

Our solution is a **dynamic coverage threshold** that increases gradually as the project matures.

## How It Works

The dynamic threshold is calculated based on the number of commits to the repository:

```
threshold = min(20 + (0.1 * commit_count), 80)
```

Where:
- The base threshold is 20%
- Each commit adds 0.1% to the threshold
- The maximum threshold is capped at 80%

### Examples

| Commit Count | Calculated Threshold | Applied Threshold |
|--------------|----------------------|-------------------|
| 0            | 20%                  | 20%               |
| 100          | 30%                  | 30%               |
| 300          | 50%                  | 50%               |
| 600          | 80%                  | 80%               |
| 800          | 100%                 | 80% (capped)      |

## Implementation

The dynamic threshold is implemented in our CI/CD pipeline:

1. The pipeline calculates the number of commits in the repository
2. It applies the formula to determine the current threshold
3. It measures the actual code coverage from the test results
4. If the actual coverage is below the threshold, the build fails

Here's the relevant code from our GitHub Actions workflow:

```yaml
- name: Calculate dynamic coverage threshold
  run: |
    # Get the number of commits
    COMMIT_COUNT=$(git rev-list --count HEAD)
    
    # Calculate the dynamic threshold: 20% + 0.1% per commit, capped at 80%
    THRESHOLD=$(echo "20 + 0.1 * $COMMIT_COUNT" | bc)
    if (( $(echo "$THRESHOLD > 80" | bc -l) )); then
      THRESHOLD=80
    fi
    echo "Dynamic coverage threshold: $THRESHOLD%"
    
    # Get the current coverage
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Total coverage: $COVERAGE%"
    
    # Check if coverage is below the threshold
    if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
      echo "Code coverage is below the dynamic threshold of $THRESHOLD%"
      exit 1
    fi
```

## Benefits

This approach provides several benefits:

1. **Gradual Improvement**: Teams can focus on improving coverage over time
2. **No Sudden Barriers**: New contributors aren't blocked by an immediate high threshold
3. **Continuous Progress**: Each commit should maintain or improve coverage
4. **Realistic Goals**: The maximum threshold (80%) acknowledges that 100% coverage isn't always practical or necessary

## Monitoring Progress

You can monitor the current coverage threshold and actual coverage in the GitHub Actions logs for each build.

## Exceptions

In some cases, certain files or packages might be excluded from coverage requirements:

- Generated code
- Test utilities
- External integrations that are difficult to test

These exceptions should be documented and justified.

## Future Enhancements

Potential future enhancements to this approach include:

1. **Per-package Thresholds**: Different thresholds for different packages
2. **Critical Path Coverage**: Higher thresholds for critical components
3. **Coverage Badges**: Displaying current coverage in the README
4. **Historical Tracking**: Graphing coverage over time
