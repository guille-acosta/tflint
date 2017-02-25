package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidIAMProfileDetector struct {
	*Detector
	profiles map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidIAMProfileDetector() *AwsInstanceInvalidIAMProfileDetector {
	return &AwsInstanceInvalidIAMProfileDetector{
		Detector: d,
		profiles: map[string]bool{},
	}
}

func (d *AwsInstanceInvalidIAMProfileDetector) PreProcess() {
	if d.isSkippable("resource", "aws_instance") {
		return
	}

	resp, err := d.AwsClient.ListInstanceProfiles()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, iamProfile := range resp.InstanceProfiles {
		d.profiles[*iamProfile.InstanceProfileName] = true
	}
}

func (d *AwsInstanceInvalidIAMProfileDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			iamProfileToken, err := hclLiteralToken(item, "iam_instance_profile")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			iamProfile, err := d.evalToString(iamProfileToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !d.profiles[iamProfile] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid IAM profile name.", iamProfile),
					Line:    iamProfileToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
