//! Tests for help, version, and status.

use std::process::Command;

/// Version identifies the binary and package version.
///
/// # Panics
///
/// Panics when an assertion fails.
#[test]
fn version() {
    let output = Command::new(env!("CARGO_BIN_EXE_afakt")).arg("--version").output();
    assert_eq!(
        output.map(|result| (result.status.success(), result.stdout, result.stderr)).ok(),
        Some((true, b"afakt 0.1.0\n".to_vec(), Vec::new()))
    );
}

/// Help lists the placeholder command and standard flags.
///
/// # Panics
///
/// Panics when an assertion fails.
#[test]
fn help() {
    let output = Command::new(env!("CARGO_BIN_EXE_afakt")).arg("--help").output();
    assert_eq!(
        output
            .map(|result| {
                let text = String::from_utf8_lossy(&result.stdout);
                (
                    result.status.success(),
                    text.contains("status"),
                    text.contains("--help"),
                    text.contains("--version"),
                    result.stderr.is_empty(),
                )
            })
            .ok(),
        Some((true, true, true, true, true))
    );
}

/// Status prints the placeholder response successfully.
///
/// # Panics
///
/// Panics when an assertion fails.
#[test]
fn status() {
    let output = Command::new(env!("CARGO_BIN_EXE_afakt")).arg("status").output();
    assert_eq!(
        output.map(|result| (result.status.success(), result.stdout, result.stderr)).ok(),
        Some((true, b"status is not implemented yet.\n".to_vec(), Vec::new()))
    );
}
