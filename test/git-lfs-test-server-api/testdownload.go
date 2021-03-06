package main

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/git-lfs/git-lfs/tools"
)

// "download" - all present
func downloadAllExist(oidsExist, oidsMissing []TestObject) error {
	retobjs, err := callBatchApi("download", oidsExist)

	if err != nil {
		return err
	}

	if len(retobjs) != len(oidsExist) {
		return fmt.Errorf("Incorrect number of returned objects, expected %d, got %d", len(oidsExist), len(retobjs))
	}

	var errbuf bytes.Buffer
	for _, o := range retobjs {
		_, ok := o.Rel("download")
		if !ok {
			errbuf.WriteString(fmt.Sprintf("Missing download link for %s\n", o.Oid))
		}
	}

	if errbuf.Len() > 0 {
		return errors.New(errbuf.String())
	}

	return nil
}

// "download" - all missing (test includes 404 error entry)
func downloadAllMissing(oidsExist, oidsMissing []TestObject) error {
	retobjs, err := callBatchApi("download", oidsMissing)

	if err != nil {
		return err
	}

	if len(retobjs) != len(oidsMissing) {
		return fmt.Errorf("Incorrect number of returned objects, expected %d, got %d", len(oidsMissing), len(retobjs))
	}

	var errbuf bytes.Buffer
	for _, o := range retobjs {
		link, ok := o.Rel("download")
		if ok {
			errbuf.WriteString(fmt.Sprintf("Download link should not exist for %s, was %s\n", o.Oid, link))
		}
		if o.Error == nil {
			errbuf.WriteString(fmt.Sprintf("Download should include an error for missing object %s, was %s\n", o.Oid))
		} else if o.Error.Code != 404 {
			errbuf.WriteString(fmt.Sprintf("Download error code for missing object %s should be 404, got %d\n", o.Oid, o.Error.Code))
		}
	}

	if errbuf.Len() > 0 {
		return errors.New(errbuf.String())
	}

	return nil
}

// "download" - mixture
func downloadMixed(oidsExist, oidsMissing []TestObject) error {

	existSet := tools.NewStringSetWithCapacity(len(oidsExist))
	for _, o := range oidsExist {
		existSet.Add(o.Oid)
	}
	missingSet := tools.NewStringSetWithCapacity(len(oidsMissing))
	for _, o := range oidsMissing {
		missingSet.Add(o.Oid)
	}

	calloids := interleaveTestData(oidsExist, oidsMissing)
	retobjs, err := callBatchApi("download", calloids)

	if err != nil {
		return err
	}

	count := len(oidsExist) + len(oidsMissing)
	if len(retobjs) != count {
		return fmt.Errorf("Incorrect number of returned objects, expected %d, got %d", count, len(retobjs))
	}

	var errbuf bytes.Buffer
	for _, o := range retobjs {
		link, ok := o.Rel("download")
		if missingSet.Contains(o.Oid) {
			if ok {
				errbuf.WriteString(fmt.Sprintf("Download link should not exist for %s, was %s\n", o.Oid, link))
			}
			if o.Error == nil {
				errbuf.WriteString(fmt.Sprintf("Download should include an error for missing object %s", o.Oid))
			} else if o.Error.Code != 404 {
				errbuf.WriteString(fmt.Sprintf("Download error code for missing object %s should be 404, got %d\n", o.Oid, o.Error.Code))
			}
		}
		if existSet.Contains(o.Oid) && !ok {
			errbuf.WriteString(fmt.Sprintf("Missing download link for %s\n", o.Oid))
		}

	}

	if errbuf.Len() > 0 {
		return errors.New(errbuf.String())
	}

	return nil

}

func init() {
	addTest("Test download: all existing", downloadAllExist)
	addTest("Test download: all missing", downloadAllMissing)
	addTest("Test download: mixed", downloadMixed)
}
