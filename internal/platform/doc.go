// Package platform isolates operating-system specific behaviour: attribute
// extraction, drive enumeration, and opening files with the default handler.
//
// This is the only package in the project that carries build-tagged,
// OS-specific code. Everything else is portable.
package platform
