//! The afakt command line skeleton.

#[cfg(test)]
mod tests;

use std::io::{self, Write};

use clap::{Parser, Subcommand};

/// The afakt command line.
#[derive(Debug, Parser)]
#[command(version, about = "A proof of concept CLI")]
struct Cli {
    /// The command to run.
    #[command(subcommand)]
    command: Command,
}

/// The commands of afakt.
#[derive(Clone, Copy, Debug, Subcommand)]
enum Command {
    /// Show the placeholder status.
    Status,
}

/// Parse the command line and write the command's output.
///
/// # Errors
///
/// Returns an error if writing to standard output fails.
fn main() -> io::Result<()> {
    run(Cli::parse().command, &mut io::stdout().lock())
}

/// Run a command with the supplied output writer.
///
/// # Errors
///
/// Returns an error if the writer fails.
fn run<W>(command: Command, output: &mut W) -> io::Result<()>
where
    W: Write,
{
    match command {
        Command::Status => write_line(output, "status is not implemented yet."),
    }
}

/// Write one line to the supplied writer.
///
/// # Errors
///
/// Returns an error if the writer fails.
fn write_line<W>(output: &mut W, line: &str) -> io::Result<()>
where
    W: Write,
{
    writeln!(output, "{line}")
}
