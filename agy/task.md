# Task List: gistsync

- [x] Project setup & initialization
- [x] Implement core models and hashing logic
- [x] Implement provider abstraction and GitHub Gist provider
- [x] Implement state management
- [x] Implement sync engine logic and conflict detection
    - [x] Evaluate command structure (`launch:login` vs others) <!-- id: 1 -->
    - [x] Research cross-platform autostart implementations <!-- id: 2 -->
    - [ ] Compare systemd vs .desktop entries for Linux <!-- id: 12 -->
    - [ ] Plan Windows autostart implementation <!-- id: 13 -->
    - [x] Plan integration with `init` flow <!-- id: 3 -->
- [x] Implement CLI commands (init, sync, status, remove, watch)
- [x] Implement file watcher
- [x] Implement main entry point
- [x] Unit tests for critical components
- [x] Final verification and distribution check
- [x] Implement periodic remote polling in watcher
