# Proposed Solution

In this folder is demonstrated a proposed solution to the "problem statement"
described in the parent `README` file related to the desire to encapsulate
step implementations within features or scenarios, yet produce a single
report file as a result.

## Overview

The proposed solution leverages standard `go` test scaffolding to define and
run multiple `godog` tests (e.g., each using their own `godog.TestSuite`)
for selected features or scenarios, then combine the outputs produced into
a single output file, as required in our case.

