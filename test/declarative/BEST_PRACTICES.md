### Best Practices
Following are some of the best practices to write an E to E experiment:
- An experiment yaml file can have one or more `kind: Recipe` custom resource(s)
- An experiment yaml file should be built to execute exactly one scenario
- A Recipe can depend on another Recipe via former's `spec.eligible` field
- An experiment yaml file may have only one Recipe with below label:
    - `d-testing.dope.mayadata.io/inference: "true"`
    - This label enables the Recipe to be considered by `inference.yaml`
- `inference.yaml` is used to decide if the ci passed or failed
- `inference.yaml` needs to be modified with every new test scenario
- `inference.yaml` can be used to include `failed`, `error`, `positive`, etc. scenarios
- Experiment yamls meant to run negative testing can be suffixed with `-neg`
    - Where `neg` stands for negative
- A negative test experiment will have its `status.phase` set with `Failed` by dope controller
- Presence of one or more failed test experiments will not fail the E to E
    - `inference.yaml` needs to be updated to include failed test cases
- Presence of one or more error test experiments will not fail the E to E
    - `inference.yaml` needs to be updated to include error test cases