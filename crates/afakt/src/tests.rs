//! Tests for command dispatch and output.

use std::io;

use super::{Command, run, write_line};

/// Status writes its placeholder response.
///
/// # Panics
///
/// Panics when an assertion fails.
#[test]
fn status_writes_placeholder() {
    let mut output = Vec::new();
    assert_eq!(run(Command::Status, &mut output).ok(), Some(()));
    assert_eq!(output, b"status is not implemented yet.\n");
}

/// The writer adds exactly one newline.
///
/// # Panics
///
/// Panics when an assertion fails.
#[test]
fn writer_adds_newline() {
    let mut output = Vec::new();
    assert_eq!(write_line(&mut output, "hello").ok(), Some(()));
    assert_eq!(output, b"hello\n");
}

/// A full writer returns its error through command dispatch.
///
/// # Panics
///
/// Panics when an assertion fails.
#[test]
fn writer_error_propagates() {
    let mut output = [0u8; 0];
    assert_eq!(
        run(Command::Status, &mut output.as_mut_slice()).err().map(|error| error.kind()),
        Some(io::ErrorKind::WriteZero)
    );
}
