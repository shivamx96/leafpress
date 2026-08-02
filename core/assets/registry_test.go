package assets

import "testing"

// pinnedBuiltins locks the stable identifiers of the built-in set. If this
// test fails, a built-in changed — update the pin deliberately. RegistryID
// is content-derived, so it tracks any such change automatically.
var pinnedBuiltins = map[string]struct {
	sha256      string
	outputPath  string
	contentType string
}{
	BuiltinFaviconICO: {
		sha256:      "35f29c9301797bf2ce58c8a79fde16c92ec7a4a16fac131cea9b9d230b809370",
		outputPath:  "favicon.ico",
		contentType: "image/x-icon",
	},
	BuiltinFaviconSVG: {
		sha256:      "796de5b284c308091c61694d4279e5f2f8a76bbdc391e9eb3533087087a85dec",
		outputPath:  "favicon.svg",
		contentType: "image/svg+xml",
	},
	BuiltinFaviconPNG: {
		sha256:      "74ff24e72b334920a961bb8983935a37126e559fdc1d5d7a10b40afb508eed25",
		outputPath:  "favicon-96x96.png",
		contentType: "image/png",
	},
	"static/leafpress/fonts/atkinson-hyperlegible-mono-italic-latin-ext.woff2": {sha256: "05c9dc1f015daa3e4773fc34385e689e9d2c6d2405d92efea3cd34f7ef21f952"},
	"static/leafpress/fonts/atkinson-hyperlegible-mono-italic-latin.woff2":     {sha256: "13db24b4da3f549a2c960880a5220a8a89df421737a62455fbd3c54e5114ca64"},
	"static/leafpress/fonts/atkinson-hyperlegible-mono-normal-latin-ext.woff2": {sha256: "3fa1fa7e990377c87018d62f356cc42b8db39fd88cee5724431e9f0d63b7f32f"},
	"static/leafpress/fonts/atkinson-hyperlegible-mono-normal-latin.woff2":     {sha256: "2706b1ee4f452e744ea91f7e4908cbde9c5d35521bf5ffffc71a382a2de89613"},
	"static/leafpress/fonts/atkinson-hyperlegible-next-italic-latin-ext.woff2": {sha256: "c0c8da569118e5b2c17032c4f1d0e293271e478fb50d869aefd08fcffdf2042b"},
	"static/leafpress/fonts/atkinson-hyperlegible-next-italic-latin.woff2":     {sha256: "4a5037bfaf6680f40147407407ec09fa42925774bde809a579283d27f9f08106"},
	"static/leafpress/fonts/atkinson-hyperlegible-next-normal-latin-ext.woff2": {sha256: "7dae0c6c66af1aec82e096186aeb1f0e6fa36ab8061ed65422f2b7daf4dd93f3"},
	"static/leafpress/fonts/atkinson-hyperlegible-next-normal-latin.woff2":     {sha256: "18b2a1a39a2fa298b0ba5390aca68462669826c90925656f1c1f6796e0e1bbaf"},
	"static/leafpress/fonts/bricolage-grotesque-normal-latin-ext.woff2":        {sha256: "104b93499342ede4da68d37234be0e5229345f0be0b9509328f3071f5fb9e8c8"},
	"static/leafpress/fonts/bricolage-grotesque-normal-latin.woff2":            {sha256: "9fee080fcc2d2e0ea8c7ce2a58abaa8ba1f40c6e603643327cd5eb6f07db06a8"},
	"static/leafpress/fonts/crimson-pro-italic-latin-ext.woff2":                {sha256: "01df7b443a6f81915a65df47439121da8307f8a574fadd7cde57c77d0b68c77b"},
	"static/leafpress/fonts/crimson-pro-italic-latin.woff2":                    {sha256: "b3faa8f9ce36db53253e11fb107d77b983150c5c21dc8fcf3906234530ab69f2"},
	"static/leafpress/fonts/crimson-pro-normal-latin-ext.woff2":                {sha256: "ac89a08a6e022ad4f3af961eb95474b7d8531c67ff88b40194bb3569d2d64e1d"},
	"static/leafpress/fonts/crimson-pro-normal-latin.woff2":                    {sha256: "20ce4189b9e41b3439a2a36dd63deff44b6d91182532202cb96b65521b4a3c23"},
	"static/leafpress/fonts/fira-code-normal-latin-ext.woff2":                  {sha256: "801677342d1191c5e964719bbcb5834f5da3c39a00e1e1501f450b1379fcc116"},
	"static/leafpress/fonts/fira-code-normal-latin.woff2":                      {sha256: "771bf4b79a97fc005d12866168bd39d868a9dd3d5903008fe8b796723c8a56f4"},
	"static/leafpress/fonts/fraunces-italic-latin-ext.woff2":                   {sha256: "877bcc2fd1be299c950d0971578336818ca5818e1e956f12e0d8c6736436ca45"},
	"static/leafpress/fonts/fraunces-italic-latin.woff2":                       {sha256: "bceec2ef4d549efbc8df0194a8d5280b6a64c3e399244dffccd9ea1bd9ad6db7"},
	"static/leafpress/fonts/fraunces-normal-latin-ext.woff2":                   {sha256: "a21ecfbf41fbc393e24ef9b7e38532a27e8da5e0a074aa7d66802d1b5ccec2f0"},
	"static/leafpress/fonts/fraunces-normal-latin.woff2":                       {sha256: "7f9d191d999336d3b9790afa72e1358e50a13b06d4f289341e92a311967a80f9"},
	"static/leafpress/fonts/geist-italic-latin-ext.woff2":                      {sha256: "56dc9ac98e3fc45e7ee487d03ee205b35858c672f2e15053eabbb119c0bc1e74"},
	"static/leafpress/fonts/geist-italic-latin.woff2":                          {sha256: "9b10496762af92659f3b05d2b084b0c8f962c3ecdf637aa764e3b7fd17f5acaf"},
	"static/leafpress/fonts/geist-mono-italic-latin-ext.woff2":                 {sha256: "cde10355ce14bb325e89af19c7609fac1fe90ea741f03a278cb3a11b02fe7f1a"},
	"static/leafpress/fonts/geist-mono-italic-latin.woff2":                     {sha256: "7842d88c32207514dd03ddb9abc6c84b032d6fef58af1177e0b23f6554c80c75"},
	"static/leafpress/fonts/geist-mono-normal-latin-ext.woff2":                 {sha256: "1a189eb997c3e2ece68373e387afaec9e8617424186c4b1ab3cff7c54ba6223b"},
	"static/leafpress/fonts/geist-mono-normal-latin.woff2":                     {sha256: "684ad5b531f81d43c1e8c7038262d5db7cdc1f68006e04d6c7769efa8d33c8cc"},
	"static/leafpress/fonts/geist-normal-latin-ext.woff2":                      {sha256: "824f485b5d26e2f2da3c2b236132ece1bc8e4e43373452950bb0e40548b4313f"},
	"static/leafpress/fonts/geist-normal-latin.woff2":                          {sha256: "19f9c92546aa300c312235e3125af1b81394d8db9a4bc4a425cd5b641d2d54e1"},
	"static/leafpress/fonts/ibm-plex-mono-italic-latin-400.woff2":              {sha256: "0840095faae86403735a8c04014b72cb29e7923646b222b360b7e8252932e4e3"},
	"static/leafpress/fonts/ibm-plex-mono-italic-latin-700.woff2":              {sha256: "2bb0d4c82652472446ed1a5a89a0553f2db1e4ea17cbda66a45e690899d83e91"},
	"static/leafpress/fonts/ibm-plex-mono-italic-latin-ext-400.woff2":          {sha256: "1b07eee6df7d26b31864581c2365131d384077605e0b63ae67c4a0a64363a6de"},
	"static/leafpress/fonts/ibm-plex-mono-italic-latin-ext-700.woff2":          {sha256: "18c8dc28d75ecbda23c88fb11126852b2232916ec2e6d6dc0225def886a0cc98"},
	"static/leafpress/fonts/ibm-plex-mono-normal-latin-400.woff2":              {sha256: "08949f728dc52d528e69b1667d15c89a5686a4ee9a296ff90983985f99c380f7"},
	"static/leafpress/fonts/ibm-plex-mono-normal-latin-700.woff2":              {sha256: "4f84d86cfd060f4ded334358ff8a4c81d4db2ed5addd568359d693f44a87765a"},
	"static/leafpress/fonts/ibm-plex-mono-normal-latin-ext-400.woff2":          {sha256: "6bc0f226a5b7884a8170e3f62c63d7675609d4631bdc5931b5cdab81821f00eb"},
	"static/leafpress/fonts/ibm-plex-mono-normal-latin-ext-700.woff2":          {sha256: "5b9b81f54dd69635c7adcaacd4c4545a73fe4809c528e22734b238a83a74135f"},
	"static/leafpress/fonts/ibm-plex-sans-italic-latin-ext.woff2":              {sha256: "bf44b77cc751203b2f0ce95c048a9d8f984313f84b58106c7c4410ca50314016"},
	"static/leafpress/fonts/ibm-plex-sans-italic-latin.woff2":                  {sha256: "4363757d2695c7633f98472855849f7c30528eba869d7968ee35b58c87cf14c1"},
	"static/leafpress/fonts/ibm-plex-sans-normal-latin-ext.woff2":              {sha256: "d160e20920ae4d6556518d352d3af27a74e9b0de3d8fe17b1c1044fc75aa2f81"},
	"static/leafpress/fonts/ibm-plex-sans-normal-latin.woff2":                  {sha256: "e2291e842cf5af167122a22881a740c7f2dda7716f1e8cd76680264f4a859470"},
	"static/leafpress/fonts/inter-italic-latin-ext.woff2":                      {sha256: "e3f94a6c4b2177643513bf5e2317b8458183d064e40031fa04c074136be205e8"},
	"static/leafpress/fonts/inter-italic-latin.woff2":                          {sha256: "7ddd9658a57d7c5811b4f6aa43625433d72c05ac83c46a231ef3dfd88813c8a1"},
	"static/leafpress/fonts/inter-normal-latin-ext.woff2":                      {sha256: "a28eb6d3ccb534ae0c94ca999371df024aab60b08c3c8a5720ee9e32fa0faaa2"},
	"static/leafpress/fonts/inter-normal-latin.woff2":                          {sha256: "c940764593d0fe5d596be327ca7558855e018039fb78509aa21921fd3644c3e4"},
	"static/leafpress/fonts/jetbrains-mono-italic-latin-ext.woff2":             {sha256: "3c6c272187a97d8e519af98b6c1255dc02108ef26e3b4bef205e3021128d7dc2"},
	"static/leafpress/fonts/jetbrains-mono-italic-latin.woff2":                 {sha256: "e3918da535efc38ebdd4e5d2748de8ecf317e930bb74cd600e8e290ca271fc82"},
	"static/leafpress/fonts/jetbrains-mono-normal-latin-ext.woff2":             {sha256: "9c38cb2d0d2d93c1ee6e21fa78db76f13ea7e15e15cc64214c7ca89b6aaa35c4"},
	"static/leafpress/fonts/jetbrains-mono-normal-latin.woff2":                 {sha256: "2c32b9b3ee358c119e210f6f5195f9bd34894d78a785ff2e95d60e718e400af4"},
	"static/leafpress/fonts/lora-italic-latin-ext.woff2":                       {sha256: "45f989df83f3a9f40007d2ecdea98bff248c8767a1a0bbe60e94fb37b605ecee"},
	"static/leafpress/fonts/lora-italic-latin.woff2":                           {sha256: "d824d807d4d832d12c87932d0b8ec1314dcfd502157a56dee6bb04cf8a3768ae"},
	"static/leafpress/fonts/lora-normal-latin-ext.woff2":                       {sha256: "2a2d9c22c9863086a23f5013fede1428585321812b25f2662542c39d02967c5e"},
	"static/leafpress/fonts/lora-normal-latin.woff2":                           {sha256: "ddb8c66035104e233fc024669183aad3738b6daa16deee2ebb1241bd0f98ace1"},
	"static/leafpress/fonts/source-code-pro-italic-latin-ext.woff2":            {sha256: "da6214a7b9f9eca88a7ae3b7b6fec00db41741774d8a38e6a899bd8e85968411"},
	"static/leafpress/fonts/source-code-pro-italic-latin.woff2":                {sha256: "97f01332aeda642a2f298030203dd4dd72802c0288ab2705d33daf691278e563"},
	"static/leafpress/fonts/source-code-pro-normal-latin-ext.woff2":            {sha256: "ab316ee7e4a28a149fe9f74a3e9173c3d14a11bb7d4df898dd2fe10da7b7fe24"},
	"static/leafpress/fonts/source-code-pro-normal-latin.woff2":                {sha256: "8b774aaa5137a38ef40f4ac9d36db9a5eee152b2f66589dfdc82ff007fc87135"},
	"static/leafpress/fonts/source-serif-4-italic-latin-ext.woff2":             {sha256: "515639854d3566c43860d2005770645c590df8b43a0144c70fe2566c33015ede"},
	"static/leafpress/fonts/source-serif-4-italic-latin.woff2":                 {sha256: "663e7ef3037a56dce81dfc33f68c1e6445995ffd8887991b3c0b68a7689c9da5"},
	"static/leafpress/fonts/source-serif-4-normal-latin-ext.woff2":             {sha256: "41529a5b38008d9ea01e28ec18693a714a3216669ee477d83a5b9db999369625"},
	"static/leafpress/fonts/source-serif-4-normal-latin.woff2":                 {sha256: "c1df4596be5029233ed2afbb8b2f6ea20784b3fb1aa5d6b5c6519ccd85eb3dfb"},
	"static/leafpress/fonts/space-grotesk-normal-latin-ext.woff2":              {sha256: "952dddb45d2f96f71cbf3b7f510b24379afc3c89ea02fcf89d377b45d62c0166"},
	"static/leafpress/fonts/space-grotesk-normal-latin.woff2":                  {sha256: "0640890476fc1198ab4de571fb658de443c4d85b66466ec09534a8737ab1ce9d"},
	"static/leafpress/fonts/OFL-atkinson-hyperlegible-mono.txt":                {sha256: "05327050630a640eb824c8fcb7de8b2e605600d8863e4f531ca7c348239916c6"},
	"static/leafpress/fonts/OFL-atkinson-hyperlegible-next.txt":                {sha256: "474a034e99845062f661855caaa505880d0e9009deddc2727a92795eab374955"},
	"static/leafpress/fonts/OFL-bricolage-grotesque.txt":                       {sha256: "46ba5f18ee20ea529f21d96c0ef8637a8314c1a5cfb2aa84018bc8157cbeff41"},
	"static/leafpress/fonts/OFL-crimson-pro.txt":                               {sha256: "a36998c273ee904e6303921fce9da2fe49c96b42cfb0bc1b70de9c3b58546a41"},
	"static/leafpress/fonts/OFL-fira-code.txt":                                 {sha256: "126cd7bf732b86f42885f1fbf682e6aeea456672c7fdb99f2d557ec44e468a28"},
	"static/leafpress/fonts/OFL-fraunces.txt":                                  {sha256: "cd3384cafac6f2bddc3955273958a2e029f97027c8037ac539ef5744a77b579e"},
	"static/leafpress/fonts/OFL-geist-mono.txt":                                {sha256: "cc815ed4fc045f0e991abb10395b7932bd028c6a067deb13316d6002105074e6"},
	"static/leafpress/fonts/OFL-geist.txt":                                     {sha256: "71609cbb5c78b5870d712eab73a31d76622635c6ed034ab5cee3b9ecbda8685f"},
	"static/leafpress/fonts/OFL-ibm-plex-mono.txt":                             {sha256: "23b0a9d0c6d3f140a0b77e483c5cfa6bba574325ef5cb189ed9f2fec4884533f"},
	"static/leafpress/fonts/OFL-ibm-plex-sans.txt":                             {sha256: "d0283623ef57e722fd0eb688a8041589670c608ab780cd3612d06ba6f153d3fd"},
	"static/leafpress/fonts/OFL-inter.txt":                                     {sha256: "5b9321a4298cfeb6b34354164a1c3afc3db114569984c502b9b35d988fd58c57"},
	"static/leafpress/fonts/OFL-jetbrains-mono.txt":                            {sha256: "b2fe5e8987594e9ffd1d2ca52a2f5d73eb8335243893c5d6254b5ad69269591d"},
	"static/leafpress/fonts/OFL-lora.txt":                                      {sha256: "948aed765160f7b974105606c21994715a171f89fca48e57bf1e0a23aca0bef1"},
	"static/leafpress/fonts/OFL-source-code-pro.txt":                           {sha256: "18aabf190848725e2576eefb5c29ba06aac1029d02132252a7f312eac2e50cf3"},
	"static/leafpress/fonts/OFL-source-serif-4.txt":                            {sha256: "18aabf190848725e2576eefb5c29ba06aac1029d02132252a7f312eac2e50cf3"},
	"static/leafpress/fonts/OFL-space-grotesk.txt":                             {sha256: "18a4de52385f6b988782639d5d0cc1326e5a8c2de9a7f01d7b20d9aedcc60943"},
	BuiltinMermaidJS: {
		sha256:      "a43bc1afd446f9c4cc66ac5dd45d02e8d65e26fc5344ec0ef787f88d6ddb6f9e",
		contentType: "text/javascript; charset=utf-8",
	},
	BuiltinMermaidLicense: {
		sha256:      "ec9fb67dcb25eccc416ed56e1aab819222c805a2a4bfe4cb19e7556bf2ffde80",
		contentType: "text/plain; charset=utf-8",
	},
}

func TestBuiltinRegistryPinned(t *testing.T) {
	all := Builtins()
	if len(all) != len(pinnedBuiltins) {
		t.Fatalf("registry has %d builtins, pin covers %d — update pins deliberately", len(all), len(pinnedBuiltins))
	}
	for _, b := range all {
		pin, ok := pinnedBuiltins[b.Asset.LogicalPath]
		if !ok {
			t.Errorf("unpinned builtin %q — add a pin deliberately", b.Asset.LogicalPath)
			continue
		}
		if b.Asset.SHA256 != pin.sha256 {
			t.Errorf("%s: sha256 = %s, pinned %s — content changed", b.Asset.LogicalPath, b.Asset.SHA256, pin.sha256)
		}
		if b.Asset.OutputPath != pin.outputPath {
			t.Errorf("%s: outputPath = %q, pinned %q", b.Asset.LogicalPath, b.Asset.OutputPath, pin.outputPath)
		}
		if pin.contentType != "" && b.Asset.ContentType != pin.contentType {
			t.Errorf("%s: contentType = %q, pinned %q", b.Asset.LogicalPath, b.Asset.ContentType, pin.contentType)
		}
	}
}

func TestValidateUserAssetPolicy(t *testing.T) {
	valid := Asset{
		LogicalPath: "static/fonts/my.woff2",
		ContentType: "font/woff2",
		SHA256:      Sum([]byte("x")),
		Size:        1,
	}
	if err := ValidateUserAsset(valid); err != nil {
		t.Fatalf("plain user asset rejected: %v", err)
	}

	override := valid
	override.LogicalPath = "static/my-favicon.ico"
	override.ContentType = "image/x-icon"
	override.OutputPath = "favicon.ico"
	if err := ValidateUserAsset(override); err != nil {
		t.Fatalf("favicon override rejected: %v", err)
	}

	reserved := valid
	reserved.LogicalPath = BuiltinPrefix + "evil.woff2"
	if err := ValidateUserAsset(reserved); err == nil {
		t.Error("reserved-namespace logical path accepted")
	}

	arbitrary := valid
	arbitrary.OutputPath = "style.css"
	if err := ValidateUserAsset(arbitrary); err == nil {
		t.Error("non-override outputPath accepted")
	}
}

func TestRootBuiltinsAreTheOverridableSet(t *testing.T) {
	roots := RootBuiltins()
	overridable := OverridableOutputPaths()
	if len(roots) != len(overridable) || len(roots) == 0 {
		t.Fatalf("RootBuiltins (%d) and OverridableOutputPaths (%d) must describe the same non-empty set", len(roots), len(overridable))
	}
	for _, b := range roots {
		if !overridable[b.Asset.OutputPath] {
			t.Errorf("%s missing from overridable set", b.Asset.OutputPath)
		}
	}
}

func TestBuiltinRegistryConsistency(t *testing.T) {
	if err := BuiltinManifest().Validate(); err != nil {
		t.Fatalf("builtin manifest invalid: %v", err)
	}
	for _, b := range Builtins() {
		if !IsBuiltinPath(b.Asset.LogicalPath) {
			t.Errorf("%s: not under %s", b.Asset.LogicalPath, BuiltinPrefix)
		}
		content := b.Content()
		if got := Sum(content); got != b.Asset.SHA256 {
			t.Errorf("%s: content hash %s does not match asset hash %s", b.Asset.LogicalPath, got, b.Asset.SHA256)
		}
		if int64(len(content)) != b.Asset.Size {
			t.Errorf("%s: content length %d does not match size %d", b.Asset.LogicalPath, len(content), b.Asset.Size)
		}
		if len(content) == 0 {
			t.Errorf("%s: empty content", b.Asset.LogicalPath)
		}
	}
}

func TestBuiltinContentIsDefensiveCopy(t *testing.T) {
	b, ok := BuiltinByLogicalPath(BuiltinFaviconICO)
	if !ok {
		t.Fatal("favicon.ico builtin not found")
	}
	first := b.Content()
	for i := range first {
		first[i] = 0xFF
	}
	second := b.Content()
	if got := Sum(second); got != b.Asset.SHA256 {
		t.Fatal("mutating a returned Content() slice corrupted the registry")
	}
}

func TestBuiltinByLogicalPath(t *testing.T) {
	b, ok := BuiltinByLogicalPath(BuiltinFaviconSVG)
	if !ok {
		t.Fatal("favicon.svg builtin not found")
	}
	if b.Asset.ContentType != "image/svg+xml" {
		t.Errorf("favicon.svg content type = %q", b.Asset.ContentType)
	}
	if _, ok := BuiltinByLogicalPath("static/leafpress/nope"); ok {
		t.Error("lookup of unknown path succeeded")
	}
}

func TestRegistryIDDerivedFromManifest(t *testing.T) {
	id := RegistryID()
	if len(id) != 64 {
		t.Fatalf("RegistryID %q is not a sha256 hex digest", id)
	}
	if id != RegistryID() {
		t.Error("RegistryID not stable across calls")
	}
	// Derived from the canonical manifest: any change to a built-in changes
	// the manifest and therefore the ID — no version integer to forget.
	if id == Sum([]byte("[]")) {
		t.Error("RegistryID suspiciously matches an empty manifest")
	}
}
