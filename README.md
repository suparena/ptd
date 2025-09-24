# PTD - Portable Tournament Data

A universal data exchange format for tournament management systems.

## Overview

PTD (Portable Tournament Data) is a standardized format for exchanging tournament data between different systems. It provides:

- **Universal Schema**: Common data model for tournaments, events, matches, and players
- **Envelope Pattern**: Consistent wrapper with metadata and versioning
- **Package Format**: ZIP archives with manifest and integrity checking
- **Digital Signatures**: Ed25519 signatures for authenticity (coming soon)
- **Vendor Extensions**: Flexibility for system-specific data

## Installation

```bash
go get github.com/suparena/ptd
```

## Quick Start

### Creating PTD Entities

```go
package main

import (
    "github.com/suparena/ptd"
    "time"
)

func main() {
    // Create a tournament
    tournament := ptd.Tournament{
        Name:      "Summer Championship 2025",
        StartDate: time.Now().Add(30 * 24 * time.Hour),
        EndDate:   time.Now().Add(32 * 24 * time.Hour),
        Status:    "published",
        Format:    "round_robin",
    }

    // Wrap in envelope
    envelope := ptd.Envelope[ptd.Tournament]{
        ID:   ptd.GenerateID(ptd.TypeTournament),
        Type: ptd.TypeTournament,
        Spec: tournament,
        Meta: ptd.Meta{
            Schema:    "ptd.v1.tournament@1.0.0",
            Version:   1,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
            Source:    "icc:prod",
        },
    }

    // Validate
    if err := envelope.Validate(); err != nil {
        panic(err)
    }
}
```

### Creating a PTD Package

```go
// Create a new package
pkg := ptd.NewPackage("Summer Championship Export")

// Add tournaments
tournaments := []interface{}{
    envelope1,
    envelope2,
}
pkg.AddEntities("tournament", tournaments)

// Add events
events := []interface{}{
    eventEnvelope1,
    eventEnvelope2,
}
pkg.AddEntities("event", events)

// Create archive
if err := pkg.CreateArchive("tournament_export.ptd"); err != nil {
    panic(err)
}
```

### Reading a PTD Package

```go
// Open package
pkg, err := ptd.OpenPackage("tournament_export.ptd")
if err != nil {
    panic(err)
}

// Access manifest
fmt.Printf("Package created: %s\n", pkg.Manifest.Created)
fmt.Printf("Total files: %d\n", len(pkg.Manifest.Files))

// Check entity counts
for entityType, count := range pkg.Manifest.Entities {
    fmt.Printf("%s: %d entities\n", entityType, count.Count)
}
```

## Entity Types

### Core Entities

- **Tournament**: Top-level competition container
- **Event**: Competition category within a tournament
- **Match**: Individual match between entries
- **Entry**: Participant registration in an event
- **Player**: Individual player information
- **Score**: Match result information

### Supporting Types

- **Venue**: Competition location
- **Organizer**: Tournament organizer
- **AgeGroup**: Age categories
- **Rules**: Competition rules
- **Rating**: Player ratings

## Package Structure

A PTD package is a ZIP archive with the following structure:

```
tournament_export.ptd
├── manifest.json           # Package manifest with checksums
├── tournament/
│   └── tournaments.ndjson # Tournament entities (newline-delimited JSON)
├── event/
│   └── events.ndjson      # Event entities
├── match/
│   └── matches.ndjson     # Match entities
├── entry/
│   └── entries.ndjson     # Entry entities
└── signature.sig          # Digital signature (optional)
```

## Metadata

Every PTD entity includes metadata:

```json
{
  "meta": {
    "schema": "ptd.v1.tournament@1.0.0",
    "version": 1,
    "created_at": "2025-09-24T12:00:00Z",
    "updated_at": "2025-09-24T12:00:00Z",
    "source": "icc:prod",
    "tags": ["official", "rated"],
    "signature": {
      "algorithm": "ed25519",
      "public_key_id": "suparena-master-2025",
      "signature": "base64_signature_here"
    }
  }
}
```

## Roadmap

- [x] Core entity types
- [x] Envelope format
- [x] Package creation/reading
- [ ] Digital signatures (Ed25519)
- [ ] Schema validation
- [ ] Import/Export handlers
- [ ] Governance framework
- [ ] Analytics extensions

## Contributing

PTD is an open standard. Contributions and feedback are welcome!

## License

MIT License - See LICENSE file for details

## Contact

Suparena Software Inc.
- Website: https://suparena.com
- Email: support@suparena.com
