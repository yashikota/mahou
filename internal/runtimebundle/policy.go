package runtimebundle

import (
	"os"
	"path/filepath"
)

func ApplyPolicy(root string, permissive bool) error {
	dir := filepath.Join(root, "etc", "ImageMagick-7")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	policy := safePolicyXML
	if permissive {
		policy = permissivePolicyXML
	}
	return os.WriteFile(filepath.Join(dir, "policy.xml"), []byte(policy), 0o644)
}

const safePolicyXML = `<?xml version="1.0" encoding="UTF-8"?>
<policymap>
  <policy domain="coder" rights="none" pattern="PDF" />
  <policy domain="coder" rights="none" pattern="PS" />
  <policy domain="coder" rights="none" pattern="PS2" />
  <policy domain="coder" rights="none" pattern="PS3" />
  <policy domain="coder" rights="none" pattern="EPS" />
  <policy domain="coder" rights="none" pattern="EPS2" />
  <policy domain="coder" rights="none" pattern="EPS3" />
  <policy domain="coder" rights="none" pattern="MVG" />
  <policy domain="coder" rights="none" pattern="MSL" />
  <policy domain="delegate" rights="none" pattern="URL" />
  <policy domain="delegate" rights="none" pattern="HTTP" />
  <policy domain="delegate" rights="none" pattern="HTTPS" />
</policymap>
`

const permissivePolicyXML = `<?xml version="1.0" encoding="UTF-8"?>
<policymap>
</policymap>
`
