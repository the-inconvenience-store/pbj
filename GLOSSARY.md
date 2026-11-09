# Clipboard

Vocabulary for enriched copying and pasting on macOS.

## Language

**Clipboard**:
The macOS pasteboard holding content for transfer between applications.

**Clipboard item**:
One copied object offered in one or more alternative representations. A clipboard may contain several items, such as several file references.

**Representation**:
One format in which copied content is offered to a receiving application, identified by a pasteboard type.
_Avoid_: Conversion, when only the advertised type changes

**Pasteboard type**:
An identifier describing a representation's format, including application-specific formats that need not be registered with macOS.
_Avoid_: File extension, MIME type

**Format conversion**:
A transformation of content from one format into another, distinct from advertising unchanged content under a different pasteboard type.

**Plain-text fallback**:
An additional plain-text representation of copied content for receiving applications that cannot use its richer representation.

**File reference**:
Copied content identifying an existing filesystem location, such as a file, directory, or symlink, rather than containing that location's contents.
_Avoid_: File contents

**Automatic detection**:
Recognition of an input format from its content without an explicit type supplied by the user.

**Publication**:
Replacement of the selected pasteboard's content with newly copied content.
