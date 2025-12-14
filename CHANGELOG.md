# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2024-12-14

### Added
- Multi-provider support: DeepSeek, Google Gemini, Alibaba DashScope
- Provider selection via CLI flags (`--pro-provider`, `--con-provider`, `--judge-provider`)
- Model selection per role (`--pro-model`, `--con-model`, `--judge-model`)
- Google Gemini official SDK integration
- Context cancellation support (Ctrl+C graceful shutdown)

### Changed
- Refactored LLM client to use interface-based design
- Default judge provider changed from OpenRouter to Gemini

## [0.1.0] - 2024-12-14

### Added
- Initial release
- Multi-persona debate workflow (Affirmative, Negative, Adjudicator)
- OpenRouter-compatible LLM client with streaming support
- CLI with file, stdin, and interactive input modes
- Configurable models for each debate role
- Colored terminal output with progress indicators
