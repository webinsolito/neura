# Exact-SHA validation trigger

This candidate exists only to trigger the pull-request-only V1 validation suite from the current `main` tree. It makes no runtime, dependency, configuration, or product behavior changes.

Base main SHA when created: `efde1e067ee449f3882502e51f78a36b2c49063f`.

Promotion rule: this marker must not be treated as a product change; validation evidence applies to the candidate tree, whose runtime content is otherwise identical to that base.
