# Changelog
All notable changes to this project will be documented in this file. See [conventional commits](https://www.conventionalcommits.org/) for commit guidelines.

- - -
## v0.2.0 - 2026-11-10T01:49:33

### Features

- **paste:** copy clipboard file references to disk ([ba74a5b](https://github.com/the-inconvenience-store/pbj/commit/ba74a5bee37b684f8c9e91928608d2810e9a07f2))

- **copy:** materialize stdin as a file attachment ([be53684](https://github.com/the-inconvenience-store/pbj/commit/be53684f31d7ff1d9584ed821873377bf47fd6a8))

- **copy:** detect HTML JSON and XML input ([06b823c](https://github.com/the-inconvenience-store/pbj/commit/06b823c193d4b48d8bb9ffeccdd99beaecb7a124))

### Bug Fixes

- **paste:** resolve symlink parents before traversal ([1205c27](https://github.com/the-inconvenience-store/pbj/commit/1205c27f83a46290f5f891fb0be2ac0eac69a5f6))

- **copy:** reject malformed XML token structures ([4176a14](https://github.com/the-inconvenience-store/pbj/commit/4176a14c4798656e56c6e59c15afd26ee146fc0e))

### Documentation

- license pbj under MIT ([bb8b45a](https://github.com/the-inconvenience-store/pbj/commit/bb8b45ac559d91b113ddedff729c5e49c1a486ee))

- record verified v0.1.0 release and Homebrew publication ([cb2ca93](https://github.com/the-inconvenience-store/pbj/commit/cb2ca93169b14a4527cf6b3b92f9df00882b740c))

### Tests

- **paste:** record filesystem and extraction CLI proof ([16da8d1](https://github.com/the-inconvenience-store/pbj/commit/16da8d10bac2c6b9876498790cce5f335e42b345))

- **copy:** record attachment CLI and Mail verification ([920d256](https://github.com/the-inconvenience-store/pbj/commit/920d256dd29baa2a72c2b1f8b2a022c6520047cb))

- **copy:** record structured input CLI verification ([717c0de](https://github.com/the-inconvenience-store/pbj/commit/717c0dee7e2961b81c8c4c0e4d3bb45f86fbf89a))


- - -

## v0.1.0 - 2026-11-09T08:51:01

### Features

- **Breaking:** rename executables to pbjcopy and pbjpaste ([67335d3](https://github.com/the-inconvenience-store/pbj/commit/67335d30c3a97b8aeb195ac15de8ed5b17e796e4))

- publish source-built Homebrew formulas without duplicate commits ([6b67e12](https://github.com/the-inconvenience-store/pbj/commit/6b67e127eacd722e62ea68b26604229e7bf60833))

- add author-free release notes and source formula rendering ([471f9ba](https://github.com/the-inconvenience-store/pbj/commit/471f9ba2ff6ca94b35240c6e6005efd34b29725e))

- add enriched pbjp extraction with Cobra ([6c04734](https://github.com/the-inconvenience-store/pbj/commit/6c04734f6e48ca561becc250098ccb0159cf87c7))

- add selective pasteboard capture and locale output encoding ([8006311](https://github.com/the-inconvenience-store/pbj/commit/8006311bbb6c6e527d71c78376418451ad985363))

- implement enriched pbjc copying with Cobra ([01b185b](https://github.com/the-inconvenience-store/pbj/commit/01b185b5287aa427d3f2289b10dfc12f28cce865))

- add native macOS pasteboard publication ([e48d985](https://github.com/the-inconvenience-store/pbj/commit/e48d985514453064b217e27ea323f2269bd55943))

### Bug Fixes

- test unlinked Homebrew installs and clean owned module caches ([1c93d9f](https://github.com/the-inconvenience-store/pbj/commit/1c93d9f2ed5bd714f8c07a164111b25dd7e10ff0))

- isolate Homebrew tap trust during source verification ([fe569a6](https://github.com/the-inconvenience-store/pbj/commit/fe569a6dc4a5f7b477e82233c8de2485fa7b6001))

- accept the source archive root directory in release verification ([33a3313](https://github.com/the-inconvenience-store/pbj/commit/33a331306aa3dcf6eb78ed1c74f909c82ce4770e))

- publish the committed Cog section for each exact release tag ([9a4ba8f](https://github.com/the-inconvenience-store/pbj/commit/9a4ba8fcfc938bba7685b6cdf83062f7d31c175e))

- allow snapshots before the first Cog changelog ([0647aa2](https://github.com/the-inconvenience-store/pbj/commit/0647aa29796b8fb2bf9f36f487459a2dd46258ab))

- reap release verifier process groups on interruption ([07e374e](https://github.com/the-inconvenience-store/pbj/commit/07e374ee86c0bd40d36f7d4a7eb91da216660c6b))

- acknowledge clipboard handoff before exiting ([5378aca](https://github.com/the-inconvenience-store/pbj/commit/5378acadad6f69a6399625c2c2d3a50fcac5f29b))

### Documentation

- record verified pbjcopy and pbjpaste cutover ([9a03136](https://github.com/the-inconvenience-store/pbj/commit/9a031368696519032d43c9f1bbfab341f4ee9a69))

- record verified source builds and release publication limits ([7df51ec](https://github.com/the-inconvenience-store/pbj/commit/7df51ec29c17bfc40ef4fe43ad5837275147e4b8))

- use pbjc and pbjp command names without bin prefixes ([6efc03c](https://github.com/the-inconvenience-store/pbj/commit/6efc03cdc2bd03ad9f60806e5e42cc5113470d7a))

### Tests

- record verified pbjp public clipboard behavior ([12b2c07](https://github.com/the-inconvenience-store/pbj/commit/12b2c077a34def483b2e40c4ddadc7af3c0db934))

- prove clipboard persistence and native application paste ([331b484](https://github.com/the-inconvenience-store/pbj/commit/331b484b4e0b1fc2bf9fb89ca29d701db60f026c))

### Continuous Integration

- release Cog tags and update source-built Homebrew tap ([76c6bbb](https://github.com/the-inconvenience-store/pbj/commit/76c6bbbd0193c2e38739f56c9cbe7db49f294812))

### Miscellaneous Chores

- migrate docs and verification to renamed commands ([e0883a4](https://github.com/the-inconvenience-store/pbj/commit/e0883a4e41aa32c041469d21da8d88062b1d900c))

- update readme ([2cbec46](https://github.com/the-inconvenience-store/pbj/commit/2cbec467e2f810050a164e3ebb6b793fc7be87ef))

- bootstrap repo ([9428e42](https://github.com/the-inconvenience-store/pbj/commit/9428e426e14003af5141f102a77dc1b07d1e35b0))


- - -

Changelog generated by [cocogitto](https://github.com/cocogitto/cocogitto).