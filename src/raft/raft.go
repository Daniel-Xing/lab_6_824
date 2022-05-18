package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

// for Lab 2A, some information:
//		time required: 1. heartbeat RPCs no more than ten times per second
//									 2. elect a new leader within five seconds
//									 3. election timeout due to the limitaion of frequency heartbeat RPC.

import (
	//	"bytes"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// role defination
const follower int = 1
const candidate int = 2
const leader int = 3

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// State
	role        int      // 1: follower 2: candidate 3: leader
	currentTerm int      // latest term server has seen
	voteFor     int      // index of peers, candidatedId that received vote in current term
	leaderId    int      // leaderId == me ? decide if me is leader
	Log         []*Entry // contain commands with log index and term

	// volatile state
	commitIndex int // index of highest log entry known to be commited
	lastApplied int //

	// volatile state for leaders
	nextIndex  []int
	matchIndex []int

	// when follower received AppendEntries rpc, send a message to channel. TODO: maybe block.
	heartbeat chan string
}

type Entry struct {
	command interface{}
	term    int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	// var term int
	// var isleader bool
	// Your code here (2A).

	rf.mu.Lock()
	defer rf.mu.Unlock()

	return rf.currentTerm, rf.me == rf.commitIndex
}

// state transfer
// beLeader clear the nextIndex and matchIndex
func (rf *Raft) beLeaer() {
	// Your code here (2B).
	rf.role = leader
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
}

// beCandidate
func (rf *Raft) beCandidate() {

	rf.role = candidate
}

// beFollower
func (rf *Raft) beFollower() {

	rf.role = follower
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
//
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

//
type AppendEntriesArgs struct {
	Term         int // leader's term
	LeaderId     int // follower can redirect client
	PrecLogIndex int // index of log entry immediately preceding new ones

	PrevLogTerm int     // term of preLogIndex entry
	Entries     []Entry // batch to store commands

	LeaderCommit int // leader's commitIndex
}

//
type AppendEntriesReply struct {
	Term    int  // currentTerm, for leader to update itself
	Success bool // true if follower matching
}

//
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {

}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int // candidate's Term
	CandidateId  int // candidate requesting vote
	LastLogIndex int // index of candidate's last log entry
	LastLogTerm  int // term of candidate's last log entry
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int  // current Term, for candidate to update itsekf
	VoteGranted bool // true means candidate received vote
}

// startElection
func (rf *Raft) startElection() {

	DPrintf("%d start election", rf.me)
	rf.mu.Lock()

	// change the role from follower to candidate
	if rf.role == follower {
		rf.beCandidate()
	}

	// currentTerm add more one
	rf.currentTerm++
	// vote for himself
	rf.voteFor = rf.me
	// pack the args
	args := &RequestVoteArgs{}
	args.Term = rf.currentTerm
	args.LastLogTerm = rf.Log[len(rf.Log)-1].term
	args.LastLogIndex = len(rf.Log) - 1

	rf.mu.Unlock()

	// init result reply
	reply := make([]*RequestVoteReply, len(rf.peers)-1)
	result := make(chan bool, len(rf.peers)-1)

	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}

		go func(result chan bool, index int, args *RequestVoteArgs, reply *RequestVoteReply) {
			ok := rf.sendRequestVote(index, args, reply)
			result <- ok
		}(result, i, args, reply[i])

	}

	voteCount := 0
	maxTerm := 0
	for i := 0; i < len(rf.peers)-1; i++ {
		if !<-result {
			continue
		}

		// find the max term from replys
		if reply[i].Term > maxTerm {
			maxTerm = reply[i].Term
		}

		if reply[i].VoteGranted {
			voteCount++
		}
	}

	rf.mu.Lock()
	// change if back to follower
	if maxTerm > rf.currentTerm {
		rf.currentTerm = maxTerm
		rf.beFollower()
	}

	// win the election and change role to the leader
	if voteCount > len(rf.peers)/2 {
		rf.beLeaer()
	}
	rf.mu.Unlock()
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	// pack the args
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// TODO: not too sure about log index and term index
	if args.Term < rf.currentTerm || args.LastLogTerm < rf.Log[len(rf.Log)-1].term ||
		args.LastLogIndex < len(rf.Log)-1 ||
		rf.voteFor != -1 || rf.voteFor != args.CandidateId {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		DPrintf("Vote rejected: %d, %d, %d, %d, %d, %d",
			args.Term, rf.currentTerm, args.LastLogTerm, args.LastLogIndex, rf.voteFor, args.CandidateId)

		return
	}

	// vote for the candidate
	reply.VoteGranted = true
	reply.Term = rf.currentTerm
	rf.voteFor = args.CandidateId
	DPrintf("Vote accepted: %d, %d, %d, %d, %d, %d",
		args.Term, rf.currentTerm, args.LastLogTerm, args.LastLogIndex, rf.voteFor, args.CandidateId)
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for rf.killed() == false {
		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().

		// leader send heartbeat to followers
		if rf.role == leader {

			time.Sleep(time.Second / 10)
			continue
		}

		// follower starts a new election or process the heartbeat
		electionTimeout := RandomTime()
		select {
		case <-time.After(electionTimeout):
			// start election
			DPrintf("Machine: %d, election happend timeout.", rf.me)
			rf.startElection()
		case msg := <-rf.heartbeat:
			// print the messages
			DPrintf("Machine: %d get the message %s", rf.me, msg)
			// process the message
		}
	}
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.currentTerm = 0
	// when init, no vote for any machine, so init as -1
	rf.voteFor = -1

	rf.Log = make([]*Entry, 0)
	rf.commitIndex = -1

	rf.leaderId = -1
	rf.lastApplied = -1

	rf.nextIndex = []int{}
	rf.matchIndex = []int{}

	rf.beFollower()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}
