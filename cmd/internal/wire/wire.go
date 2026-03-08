package wire

// Init wires dependency-injection callbacks between moslib packages.
// Every binary (mos, mgov, mvcs, mgate, mtrace, mstore) must call this
// before executing any command.
//
// With the DSL, artifact, and linter packages removed, Init is a no-op.
func Init() {}
