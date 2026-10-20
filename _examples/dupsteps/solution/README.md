# Proposed Solution

In this folder is demonstrated a proposed solution to the "problem statement"
described in the parent `README` file related to the desire to encapsulate
Step implementations within Features or Scenarios, yet produce a single
report file as a result.

## Overview

The proposed solution leverages standard `go` test scaffolding to define and
run multiple `godog` tests (e.g., each using their own `godog.TestSuite`)
for selected Features or Scenarios, then combine the outputs produced into
a single report file, as required in our case.

