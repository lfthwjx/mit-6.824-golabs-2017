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

import (
	"labrpc"
	"math"
	"math/rand"
	"sync"
	"time"
)

// import "bytes"
// import "encoding/gob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

type Role int

const (
	follower Role = iota
	candidate
	leader
)

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	term int
	// -1 if do not vote yet
	votedFor    int
	role        Role
	logs        []LogEntry
	alive       chan bool
	votes       map[int]bool
	killed      bool
	applyCh     chan ApplyMsg
	nextIndex   []int
	matchIndex  []int
	syncCh      chan bool
	wakeApply   chan bool
	commitIndex int
	lastApplied int
}

type LogEntry struct {
	Term    int
	Command interface{}
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.term
	isleader = rf.role == leader
	DPrintf("GetState(): term=%v, raftid=%v, isleader=%v", term, rf.me, isleader)
	return term, isleader
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
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Term = rf.term
	reply.Success = true
	if args.Term < rf.term {
		reply.Success = false
		return
	}
	if args.Term > rf.term {
		rf.role = follower
		rf.term = args.Term
	}
	rf.alive <- true

	if args.PrevLogIndex >= len(rf.logs) || (args.PrevLogIndex > -1 && args.PrevLogTerm != rf.logs[args.PrevLogIndex].Term) {
		reply.Success = false
		return
	}
	i := args.PrevLogIndex + 1
	j := 0
	for i < len(rf.logs) && j < len(args.Entries) {
		rf.logs[i] = args.Entries[j]
		i++
		j++
	}
	if j < len(args.Entries) {
		rf.logs = append(rf.logs, args.Entries[j:]...)
	}
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = minInt(args.LeaderCommit, len(rf.logs)-1)
		go func() {
			rf.wakeApply <- true
		}()
	}
	if len(args.Entries) > 0 {
		DPrintf("<-AppendEntries: raftid=%v, leaderCommit=%v, rf.commitIndex=%v, rf.logs=%+v, args=%+v",
			rf.me, args.LeaderCommit, rf.commitIndex, rf.logs, args)
	}
}

func (rf *Raft) applyLog() {
	rf.mu.Lock()
	for !rf.killed {
		if rf.lastApplied < rf.commitIndex {
			rf.lastApplied++
			li := rf.lastApplied
			msg := ApplyMsg{
				Index:   li,
				Command: rf.logs[li].Command,
			}
			DPrintf("applyLog: msg=%+v,term=%v,raftid=%v", msg, rf.term, rf.me)
			rf.mu.Unlock()
			rf.applyCh <- msg
			rf.mu.Lock()
			continue
		}
		rf.mu.Unlock()
		select {
		case <-rf.wakeApply:
		case <-time.After(time.Second):
		}
		rf.mu.Lock()
	}
	rf.mu.Unlock()
}

// sendAppendEntries calls rpc, retry indefinitely if rpc failed.
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	term := rf.term
	for term == rf.term && !rf.killed {
		rf.mu.Unlock()
		ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
		DPrintf("from %v: %v=Raft.AppendEntries(%+v, %+v)", server, ok, args, reply)
		rf.mu.Lock()
		if !ok {
			continue
		}
		if reply.Term > rf.term {
			rf.term = reply.Term
			rf.role = follower
		}
		break
	}
	rf.mu.Unlock()
}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	reply.VoteGranted = false
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term < rf.term {
		reply.Term = rf.term
		return
	}
	reply.Term = args.Term
	if args.Term > rf.term {
		rf.role = follower
		rf.term = args.Term
		rf.votedFor = -1
	}
	lastIndex := len(rf.logs) - 1
	lastTerm := -1
	if lastIndex > 0 {
		lastTerm = rf.logs[lastIndex].Term
	}
	DPrintf("RequestVote(%+v, _): raftid=%v votedFor=%v, lastIndex=%v, lastTerm=%v", args, rf.me, rf.votedFor, lastIndex, lastTerm)
	if (rf.votedFor == -1 || rf.votedFor == args.CandidateID) && ((args.LastLogIndex == -1 && lastIndex == -1) ||
		(lastIndex > -1 && (rf.logs[lastIndex].Term < args.LastLogTerm || (rf.logs[lastIndex].Term == args.LastLogTerm && lastIndex <= args.LastLogIndex)))) {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateID
		rf.alive <- true
		return
	}
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
	rf.mu.Lock()
	term := rf.term
	for {
		if rf.killed || rf.term > term {
			rf.mu.Unlock()
			break
		}
		rf.mu.Unlock()
		ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
		DPrintf("voteResult from %v: %v=Raft.RequestVote(%+v, %+v)", server, ok, args, reply)
		if ok {
			break
		}
		time.Sleep(time.Millisecond * 20)
		rf.mu.Lock()
	}
	return true
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
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
	// TODO think about start a command twice
	rf.mu.Lock()
	isLeader = rf.role == leader
	if !isLeader {
		rf.mu.Unlock()
		return index, term, isLeader
	}
	entries := []LogEntry{LogEntry{Term: rf.term, Command: command}}
	rf.logs = append(rf.logs, entries...)
	term = rf.term
	index = len(rf.logs) - 1
	rf.mu.Unlock()
	go func() {
		rf.syncCh <- true
	}()
	if !isLeader {
		return index, term, isLeader
	}
	return index, term, isLeader
}

// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
	rf.mu.Lock()
	rf.killed = true
	rf.mu.Unlock()
}

// startRole starts the role state machine
func (rf *Raft) startRole() {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	for {
		if rf.killed {
			return
		}
		// heartbeat every 110ms
		// election timeout: 300--500ms, for 2 heartbeat fail
		t := time.Millisecond * time.Duration(300+r1.Intn(200))
		select {
		case <-rf.alive:
			continue
		case <-time.After(t):
			// TODO for leader, we may return, but must start when we change to other role.
			if rf.killed {
				return
			}
			rf.mu.Lock()
			switch rf.role {
			case candidate, follower:
				// -> candidate
				rf.role = candidate
				rf.term++
				rf.votedFor = rf.me
				rf.votes = make(map[int]bool)
				rf.votes[rf.me] = true
				DPrintf("Election timeout: term=%v, votedFor=%v", rf.term, rf.votedFor)
				// start election
				for i := range rf.peers {
					if i == rf.me {
						continue
					}
					lt := -1
					if len(rf.logs) > 0 {
						lt = rf.logs[len(rf.logs)-1].Term
					}
					args := RequestVoteArgs{
						Term:         rf.term,
						CandidateID:  rf.me,
						LastLogIndex: len(rf.logs) - 1,
						LastLogTerm:  lt,
					}
					go func(i, term int, args RequestVoteArgs) {
						var reply RequestVoteReply
						rf.sendRequestVote(i, &args, &reply)
						if !reply.VoteGranted {
							return
						}
						rf.mu.Lock()
						defer rf.mu.Unlock()
						if rf.role != candidate || rf.term > term {
							return
						}
						rf.votes[i] = true
						if len(rf.votes) < len(rf.peers)/2+1 {
							return
						}
						if len(rf.votes) == len(rf.peers)/2+1 {
							DPrintf("Become Leader: raftid=%v, term=%v", rf.me, term)
							rf.role = leader
							rf.nextIndex = make([]int, len(rf.peers))
							rf.matchIndex = make([]int, len(rf.peers))
							for i := range rf.peers {
								if i == rf.me {
									rf.nextIndex[i] = -1
									rf.matchIndex[i] = -1
									continue
								}
								rf.nextIndex[i] = len(rf.logs)
							}
							go rf.syncLog()
							go rf.Start(nil)
							go rf.heartbeatPeriod()
						}
					}(i, rf.term, args)
				}
			}
			rf.mu.Unlock()
		}
	}
}

func (rf *Raft) syncLog() {
	rf.mu.Lock()
	DPrintf("->syncLog(): raftid=%v, term=%v", rf.me, rf.term)
	rf.mu.Unlock()
	var syncCh []chan bool
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		ch := make(chan bool)
		syncCh = append(syncCh, ch)
		go func(i int, ch chan bool) {
			rf.mu.Lock()
			term := rf.term
			for !rf.killed && term == rf.term && rf.role == leader {
				if len(rf.logs)-1 >= rf.nextIndex[i] {
					li := rf.nextIndex[i] - 1
					lt := -1
					if li > -1 {
						lt = rf.logs[li].Term
					}
					newIndex := len(rf.logs)
					DPrintf("syncLog: term=%v, currentTerm=%v, raftid=%v, peerid=%v, nextIndex=%v, %+v",
						term, rf.term, rf.me,
						i, rf.nextIndex[i], rf.logs)
					args := AppendEntriesArgs{
						Term:         rf.term,
						LeaderID:     rf.me,
						PrevLogIndex: li,
						PrevLogTerm:  lt,
						Entries:      rf.logs[rf.nextIndex[i]:newIndex],
						LeaderCommit: rf.commitIndex,
					}
					rf.mu.Unlock()
					var reply AppendEntriesReply
					rf.sendAppendEntries(i, &args, &reply)
					rf.mu.Lock()
					if !reply.Success {
						if term == rf.term && rf.role == leader {
							rf.nextIndex[i]--
							continue
						}
					} else {
						DPrintf("syncLog success: follower id=%v, nextIndex=%v", i, newIndex)
						rf.nextIndex[i] = newIndex
						rf.matchIndex[i] = newIndex - 1
						go rf.applyMaybe()
					}
				}
				rf.mu.Unlock()
				select {
				case <-ch:
				case <-time.After(time.Millisecond * 200):
				}
				rf.mu.Lock()
			}
			rf.mu.Unlock()
		}(i, ch)
	}
	rf.mu.Lock()
	term := rf.term
	for !rf.killed && term == rf.term && rf.role == leader {
		rf.mu.Unlock()
		select {
		case <-rf.syncCh:
			for _, ch := range syncCh {
				go func(ch chan bool) {
					ch <- true
				}(ch)
			}
		case <-time.After(time.Millisecond * 200):
		}
		rf.mu.Lock()
	}
	rf.mu.Unlock()
	DPrintf("<-syncLog(): raftid=%v, term=%v", rf.me, term)
}

// applyMaybe appies log if possible. must be called after rf.mu.Lock().
func (rf *Raft) applyMaybe() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	minM := rf.commitIndex
	for i := len(rf.logs) - 1; i > rf.commitIndex; i-- {
		if rf.logs[i].Term < rf.term {
			break
		}
		c := countGreater(rf.matchIndex, i)
		// len(rf.matchIndex)/2 + leader self = majority
		if c >= len(rf.matchIndex)/2 && rf.logs[i].Term == rf.term {
			minM = i
			break
		}
	}
	if minM > rf.commitIndex {
		rf.commitIndex = minM
		go func() {
			rf.wakeApply <- true
		}()
	}
}

func countGreater(vn []int, p int) int {
	c := 0
	for _, v := range vn {
		if v >= p {
			c++
		}
	}
	return c
}

func minInt(vn ...int) int {
	m := math.MaxInt64
	for i := 0; i < len(vn); i++ {
		if vn[i] < m {
			m = vn[i]
		}
	}
	return m
}

func (rf *Raft) heartbeat(term, raftid int) {
	for i := range rf.peers {
		if i == raftid {
			continue
		}
		go func(i int) {
			rf.mu.Lock()
			li := len(rf.logs) - 1
			lt := -1
			if li > 0 {
				lt = rf.logs[li].Term
			}
			args := AppendEntriesArgs{
				Term:         term,
				LeaderID:     raftid,
				PrevLogIndex: li,
				PrevLogTerm:  lt,
				Entries:      []LogEntry{},
				LeaderCommit: rf.commitIndex,
			}
			rf.mu.Unlock()
			var reply AppendEntriesReply
			rf.peers[i].Call("Raft.AppendEntries", &args, &reply)
		}(i)
	}
}

func (rf *Raft) heartbeatPeriod() {
	rf.mu.Lock()
	term := rf.term
	me := rf.me
	DPrintf("->heartbeatPeriod(): raftid=%v", rf.me)
	rf.mu.Unlock()
	rf.heartbeat(term, me)

	ticker := time.NewTicker(110 * time.Millisecond)
	for range ticker.C {
		rf.mu.Lock()
		nt := rf.term
		me := rf.me
		killed := rf.killed
		rf.mu.Unlock()
		if nt > term || killed {
			break
		}
		rf.heartbeat(nt, me)
	}
	DPrintf("<-heartbeatPeriod(): raftid=%v", rf.me)
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
	rf.term = 0
	rf.votedFor = -1
	rf.role = follower
	rf.alive = make(chan bool)
	rf.killed = false
	rf.applyCh = applyCh
	rf.commitIndex = -1
	rf.lastApplied = -1
	rf.syncCh = make(chan bool)
	rf.wakeApply = make(chan bool)

	go rf.startRole()
	go rf.applyLog()
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
