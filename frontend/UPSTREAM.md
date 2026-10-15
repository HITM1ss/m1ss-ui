# Frontend provenance

Derived from https://github.com/alireza0/s-ui-frontend at commit `e4525297b002c1c3be234cc1c9695ef84be741a0`.

The customized frontend was moved from a Git submodule into this repository on 2026-10-15 so CI and releases build exactly the same source as the local M1ss-UI interface. Upstream license and notices remain in LICENSE and the source files. The previous local submodule Git history has been preserved in the parent checkout under `.git/modules/frontend` (it is local metadata and is not part of distributed source).

Dependencies and generated output (`node_modules`, `dist`) are not committed. To incorporate upstream changes, review them against the pinned source commit and retain this project's UI and branding customizations.
