////////////////////////////////////////////////////////////////////////////
//
// Summit Route End Point Protection
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.
//
/////////////////////////////////////////////////////////////////////////////


package models

// SystemSet is a collection of Systems,
// and a SystemSet may itself belong to another SystemSet
type SystemSet struct {
	ID         int64
	CustomerID int64
	Name       string
	RuleSetID  int64

	Mode int // 0 = Monitor, 1 = Enforce rules
	// One day other modes such as enforce but allow via prompt
	SystemSetID int64 // Allows recursion

	CreationDate int64
}

// System is a computer with SREPP installed on it
type System struct {
	ID           int64
	SystemSetID  int64
	SystemUUID   []byte // ID given to the agent
	AgentVersion string
	Comment      string // User defined name

	OSHumanName  string
	OSVersion    string // Can't call this Version or things break
	Manufacturer string
	Model        string
	MachineGUID  string // From  HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography\MachineGUID
	Arch         string
	MachineName  string

	FirstSeen int64
	LastSeen  int64
}

// Task records commands for an agent
type Task struct {
	ID                  int64
	SystemID            int64
	CreationDate        int64
	DeployedToAgentDate int64
	Command             string
}

// RuleSet points to a linked list of Rules
type RuleSet struct {
	ID        int64
	FirstRule int64
}

// Rule is member of the RuleSet
type Rule struct {
	ID             int64
	Description    string
	AttributeType  string
	AttributeValue string
	AllowDeny      bool // Allow = true, Deny = false

	NextRule     int64
	PreviousRule int64
}

// ExecutableFile is an exe that became a process.
type ExecutableFile struct {
	ID int64

	//
	// This data comes from the client
	//
	// File hashes
	Md5    []byte
	Sha1   []byte
	Sha256 []byte

	CodeSectionSha256 []byte // TODO
	Size              int    // Size in bytes
	IsSigned          bool
	FirstSeen         int64 // Time first seen anywhere
	ExecutionType     int   // 1 = exe, 2 = dll, 4 = sys

	UploadDate int64 // 0 if we don't have a copy

	//
	// Data inserted by worker
	//
	AnalysisDate int64 // 0 if we haven't analyzed it yet

	// Authenticode hashes
	AuthenticodeMd5    []byte
	AuthenticodeSha1   []byte
	AuthenticodeSha256 []byte

	// Resource info
	CompanyName      string
	ProductVersion   string
	ProductName      string
	FileDescription  string
	InternalName     string
	FileVersion      string
	OriginalFilename string

	Architecture int // TODO 32, 64, or 1 (arm)
}

// ProcessEvent is tied to a system
type ProcessEvent struct {
	ID               int64
	SystemID         int64
	ExecutableFileID int64
	PID              int64
	PPID             int64
	FilePath         string // TODO MAYBE Normalize into own table
	CommandLine      string // TODO MAYBE Normalize into own table
	EventTime        int64
	State            int
}

// FileToSystemMap maps executables to systems so we don't need to search through the ProcessEvent table
type FileToSystemMap struct {
	FileID   int64 // TODO Need to set unique on (FileID, SystemID)
	SystemID int64
	FilePath string // TODO MAYBE Normalize and match this up with ProcessEvent
	// FilePath is the path where it was last seen
	FirstSeen int64
	LastSeen  int64
}

// FileToSignerMap is for the one to many relationship
// of files to signers, as a file can be signed multiple times
// [one file] -> [many signers]
type FileToSignerMap struct {
	FileID   int64 // one
	SignerID int64 // many
}

// FileToCounterSignerMap I think a file can only be counter-signed once
type FileToCounterSignerMap struct {
	FileID    int64
	Timestamp int64
	SignerID  int64
}

// Signer is used to sign an ExecutableFile
type Signer struct {
	ID                               int64
	Version                          int
	Subject                          string
	SubjectShortName                 string
	SerialNumber                     []byte
	DigestAlgorithm                  string
	DigestEncryptionAlgorithm        string
	DigestEncryptionAlgorithmKeySize int
	IssuerID                         int64 // Parent in trust chain
}

// CatalogFile is catalog file that authenticates executables
type CatalogFile struct {
	ID int64

	// This data comes from the client
	FilePath  string
	Sha256    []byte
	Size      int
	FirstSeen int64 // Time first seen anywhere

	UploadDate int64 // 0 if we don't have a copy

	//
	// Data inserted by worker
	//
	AnalysisDate int64 // 0 if we haven't analyzed it yet
	SignerID     int64
}

// CertificateTrustList is a mapping that is extracted from catalog files of SignerID's to hashes.
// When a PE file or catalog comes around that causes a match, a FileToSignerMap update is made and the FileID here is updated.
type CertificateTrustList struct {
	CatalogID int64
	Hash      []byte
	HashType  string // sha1, md5, or sha256
	FileID    int64  // Will be 0 until a match is found
}

// Updates is a mapping of what versions agents can update to
type Updates struct {
	ID          int64
	VersionFrom string
	VersionTo   string
}
