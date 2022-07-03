import os
import sys

cmd_preifx = "cd ../src/raft && go test -run "
test_name = {}
test_name["2A"] = ["TestInitialElection2A", "TestReElection2A", "TestManyElections2A"]
test_name["2B"] = ["TestBasicAgree2B", "TestRPCBytes2B", "TestFailAgree2B", "TestFailNoAgree2B", "TestConcurrentStarts2B", "TestRejoin2B", "TestBackuDPrintf", "TestCount2B"]
test_name["2C"] = ["TestPersist12C", "TestPersist22C", "TestPersist32C", "TestFigure82C", "TestUnreliableAgree2C", "TestFigure8Unreliable2C", "TestReliableChurn2C", "TestUnreliableChurn2C"]
test_name["2D"] = ["TestSnapshotBasic2D", "TestSnapshotInstall2D", "TestSnapshotInstallUnreliable2D", "TestSnapshotInstallCrash2D", "TestSnapshotInstallUnCrash2D"]

def run(cmd):
    return os.popen(cmd).read()

def savefile(content, filename):
    with open(filename, "w+") as f:
        f.write(content)

def find_fail(task, times = 50):
    failed = False
    for t in range(times):
        res = run(cmd_preifx + task)
        passed = res.find("Passed")
        if passed == -1:
            print("Find the bad case with %s after run %d times"%(task, t))
            savefile(res, task+"_log_%d.txt"%(t))
            failed = True
        if failed != True and (t+1) % 10 == 0:
            print("No big deal after running %d times"%(t+1))
            
    if failed is not True:
        print("No Failed with %s after running times: %d!"%(task, times))

def find_fail_with_lab(labName, times = 50):
    tasks = test_name[labName]
    for task in tasks:
        find_fail(task, times)

if __name__ == "__main__":
    args = sys.argv
    if len(args) >= 2:
        find_fail_with_lab(args[1])
    elif len(args) >= 3:
        find_fail_with_lab(args[1], int(args[2]))
