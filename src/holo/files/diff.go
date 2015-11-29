	"os/exec"
	"syscall"
//version, similar to `diff /var/lib/holo/files/provisioned/$FILE $FILE`, but it also
	toPath := target.PathIn(common.TargetDirectory())

	fromPathToUse, err := checkFile(fromPath)
	toPathToUse, err := checkFile(toPath)
	//run git-diff to obtain the diff
	var buffer bytes.Buffer
	cmd := exec.Command("git", "diff", "--no-index", "--", fromPathToUse, toPathToUse)
	cmd.Stdout = &buffer
	cmd.Stderr = os.Stderr
	//error "exit code 1" is normal for different files, only exit code > 2 means trouble
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					err = nil
				}
	//did a relevant error occur?
	if err != nil {
		return nil, err
	//remove "index <SHA1>..<SHA1> <mode>" lines
	result := buffer.Bytes()
	rx := regexp.MustCompile(`(?m:^index .*$)\n`)
	result = rx.ReplaceAll(result, nil)
	//remove "/var/lib/holo/files/provisioned" from path displays to make it appear like we
	//just diff the target path
	if fromPathToUse == fromPath {
		fromPathQuoted := strings.TrimPrefix(regexp.QuoteMeta(fromPath), "/")
		toPathQuoted := strings.TrimPrefix(regexp.QuoteMeta(toPath), "/")
		toPathTrimmed := strings.TrimPrefix(toPath, "/")
		rx = regexp.MustCompile(`(?m:^)diff --git a/` + fromPathQuoted)
		result = rx.ReplaceAll(result, []byte("diff --git a/"+toPathTrimmed))
		rx = regexp.MustCompile(`(?m:^)diff --git a/` + toPathQuoted + ` b/` + fromPathQuoted)
		result = rx.ReplaceAll(result, []byte("diff --git a/"+toPathTrimmed+" b/"+toPathTrimmed))
		rx = regexp.MustCompile(`(?m:^)--- a/` + fromPathQuoted)
		result = rx.ReplaceAll(result, []byte("--- a/"+toPathTrimmed))
	return result, nil
func checkFile(path string) (pathToUse string, returnError error) {
	//check that files are either non-existent (in which case git-diff needs to
	//be given /dev/null instead or manageable (e.g. we can't diff directories
	//or device files)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "/dev/null", nil
		}
		return path, err
	if !common.IsManageableFileInfo(info) {
		return path, fmt.Errorf("%s is not a manageable file", path)
	}
	return path, nil