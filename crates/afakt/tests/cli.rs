//! Tests that run the built afakt binary.

// The integration target exercises the binary instead of using clap directly.
use clap as _;

#[cfg(test)]
#[path = "cli/tests.rs"]
mod tests;
