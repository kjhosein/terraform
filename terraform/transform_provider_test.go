package terraform

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/dag"
)

func TestProviderTransformer(t *testing.T) {
	mod := testModule(t, "transform-provider-basic")

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &AttachResourceConfigTransformer{Config: mod}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &MissingProviderTransformer{Providers: []string{"aws"}}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	transform := &ProviderTransformer{}
	if err := transform.Transform(&g); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformProviderBasicStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestProviderTransformer_moduleChild(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}

	{
		tf := &ImportStateTransformer{
			Targets: []*ImportTarget{
				&ImportTarget{
					Addr: addrs.RootModuleInstance.
						Child("moo", addrs.NoKey).
						ResourceInstance(
							addrs.ManagedResourceMode,
							"foo_instance",
							"qux",
							addrs.NoKey,
						),
					ProviderAddr: addrs.RootModuleInstance.
						Child("moo", addrs.NoKey).
						ProviderConfigDefault(addrs.NewLegacyProvider("foo")),
					ID: "bar",
				},
			},
		}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
		t.Logf("graph after ImportStateTransformer:\n%s", g.String())
	}

	{
		tf := &MissingProviderTransformer{Providers: []string{"foo", "bar"}}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
		t.Logf("graph after MissingProviderTransformer:\n%s", g.String())
	}

	{
		tf := &ProviderTransformer{}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
		t.Logf("graph after ProviderTransformer:\n%s", g.String())
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformProviderModuleChildStr)
	if actual != expected {
		t.Fatalf("wrong result\n\ngot:\n%s\n\nwant:\n%s", actual, expected)
	}
}

// Test providers with FQNs that do not match the typeName
func TestProviderTransformer_fqns(t *testing.T) {
	for _, mod := range []string{"fqns", "fqns-module"} {
		mod := testModule(t, fmt.Sprintf("transform-provider-%s", mod))

		g := Graph{Path: addrs.RootModuleInstance}
		{
			tf := &ConfigTransformer{Config: mod}
			if err := tf.Transform(&g); err != nil {
				t.Fatalf("err: %s", err)
			}
		}

		{
			transform := &AttachResourceConfigTransformer{Config: mod}
			if err := transform.Transform(&g); err != nil {
				t.Fatalf("err: %s", err)
			}
		}

		{
			transform := &MissingProviderTransformer{Providers: []string{"aws"}, Config: mod}
			if err := transform.Transform(&g); err != nil {
				t.Fatalf("err: %s", err)
			}
		}

		transform := &ProviderTransformer{Config: mod}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := strings.TrimSpace(g.String())
		expected := strings.TrimSpace(testTransformProviderBasicStr)
		if actual != expected {
			t.Fatalf("bad:\n\n%s", actual)
		}
	}
}

func TestCloseProviderTransformer(t *testing.T) {
	mod := testModule(t, "transform-provider-basic")

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &AttachResourceConfigTransformer{Config: mod}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &MissingProviderTransformer{Providers: []string{"aws"}}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &ProviderTransformer{}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &CloseProviderTransformer{}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformCloseProviderBasicStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestCloseProviderTransformer_withTargets(t *testing.T) {
	mod := testModule(t, "transform-provider-basic")

	g := Graph{Path: addrs.RootModuleInstance}
	transforms := []GraphTransformer{
		&ConfigTransformer{Config: mod},
		&MissingProviderTransformer{Providers: []string{"aws"}},
		&ProviderTransformer{},
		&CloseProviderTransformer{},
		&TargetsTransformer{
			Targets: []addrs.Targetable{
				addrs.RootModuleInstance.Resource(
					addrs.ManagedResourceMode, "something", "else",
				),
			},
		},
	}

	for _, tr := range transforms {
		if err := tr.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(``)
	if actual != expected {
		t.Fatalf("expected:%s\n\ngot:\n\n%s", expected, actual)
	}
}

func TestMissingProviderTransformer(t *testing.T) {
	mod := testModule(t, "transform-provider-missing")

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &AttachResourceConfigTransformer{Config: mod}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &MissingProviderTransformer{Providers: []string{"aws", "foo", "bar"}}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &ProviderTransformer{}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &CloseProviderTransformer{}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformMissingProviderBasicStr)
	if actual != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

func TestMissingProviderTransformer_grandchildMissing(t *testing.T) {
	mod := testModule(t, "transform-provider-missing-grandchild")

	concrete := func(a *NodeAbstractProvider) dag.Vertex { return a }

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &AttachResourceConfigTransformer{Config: mod}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := TransformProviders([]string{"aws", "foo", "bar"}, concrete, mod)
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		transform := &TransitiveReductionTransformer{}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformMissingGrandchildProviderStr)
	if actual != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

func TestMissingProviderTransformer_moduleChild(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}

	// We use the import state transformer since at the time of writing
	// this test it is the first and only transformer that will introduce
	// multiple module-path nodes at a single go.
	{
		tf := &ImportStateTransformer{
			Targets: []*ImportTarget{
				&ImportTarget{
					Addr: addrs.RootModuleInstance.
						Child("moo", addrs.NoKey).
						ResourceInstance(
							addrs.ManagedResourceMode,
							"foo_instance",
							"qux",
							addrs.NoKey,
						),
					ProviderAddr: addrs.RootModuleInstance.
						Child("moo", addrs.NoKey).
						ProviderConfigDefault(addrs.NewLegacyProvider("foo")),
					ID: "bar",
				},
			},
		}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		tf := &MissingProviderTransformer{Providers: []string{"foo", "bar"}}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformMissingProviderModuleChildStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

func TestMissingProviderTransformer_moduleGrandchild(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}

	// We use the import state transformer since at the time of writing
	// this test it is the first and only transformer that will introduce
	// multiple module-path nodes at a single go.
	{
		tf := &ImportStateTransformer{
			Targets: []*ImportTarget{
				&ImportTarget{
					Addr: addrs.RootModuleInstance.
						Child("a", addrs.NoKey).
						Child("b", addrs.NoKey).
						ResourceInstance(
							addrs.ManagedResourceMode,
							"foo_instance",
							"qux",
							addrs.NoKey,
						),
					ProviderAddr: addrs.RootModuleInstance.
						Child("moo", addrs.NoKey).
						ProviderConfigDefault(addrs.NewLegacyProvider("foo")),
					ID: "bar",
				},
			},
		}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		tf := &MissingProviderTransformer{Providers: []string{"foo", "bar"}}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformMissingProviderModuleGrandchildStr)
	if actual != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

func TestParentProviderTransformer(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}

	// Introduce a cihld module
	{
		tf := &ImportStateTransformer{
			Targets: []*ImportTarget{
				&ImportTarget{
					Addr: addrs.RootModuleInstance.
						Child("moo", addrs.NoKey).
						ResourceInstance(
							addrs.ManagedResourceMode,
							"foo_instance",
							"qux",
							addrs.NoKey,
						),
					ProviderAddr: addrs.RootModuleInstance.
						Child("moo", addrs.NoKey).
						ProviderConfigDefault(addrs.NewLegacyProvider("foo")),
					ID: "bar",
				},
			},
		}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Add the missing modules
	{
		tf := &MissingProviderTransformer{Providers: []string{"foo", "bar"}}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Connect parents
	{
		tf := &ParentProviderTransformer{}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformParentProviderStr)
	if actual != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

func TestParentProviderTransformer_moduleGrandchild(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}

	// We use the import state transformer since at the time of writing
	// this test it is the first and only transformer that will introduce
	// multiple module-path nodes at a single go.
	{
		tf := &ImportStateTransformer{
			Targets: []*ImportTarget{
				&ImportTarget{
					Addr: addrs.RootModuleInstance.
						Child("a", addrs.NoKey).
						Child("b", addrs.NoKey).
						ResourceInstance(
							addrs.ManagedResourceMode,
							"foo_instance",
							"qux",
							addrs.NoKey,
						),
					ProviderAddr: addrs.RootModuleInstance.
						Child("moo", addrs.NoKey).
						ProviderConfigDefault(addrs.NewLegacyProvider("foo")),
					ID: "bar",
				},
			},
		}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		tf := &MissingProviderTransformer{Providers: []string{"foo", "bar"}}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	// Connect parents
	{
		tf := &ParentProviderTransformer{}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformParentProviderModuleGrandchildStr)
	if actual != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

func TestPruneProviderTransformer(t *testing.T) {
	mod := testModule(t, "transform-provider-prune")

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &AttachResourceConfigTransformer{Config: mod}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &MissingProviderTransformer{Providers: []string{"foo"}}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &ProviderTransformer{}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &CloseProviderTransformer{}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		transform := &PruneProviderTransformer{}
		if err := transform.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformPruneProviderBasicStr)
	if actual != expected {
		t.Fatalf("bad:\n\n%s", actual)
	}
}

// the child module resource is attached to the configured parent provider
func TestProviderConfigTransformer_parentProviders(t *testing.T) {
	mod := testModule(t, "transform-provider-inherit")
	concrete := func(a *NodeAbstractProvider) dag.Vertex { return a }

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		tf := &AttachResourceConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		tf := TransformProviders([]string{"aws"}, concrete, mod)
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformModuleProviderConfigStr)
	if actual != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

// the child module resource is attached to the configured grand-parent provider
func TestProviderConfigTransformer_grandparentProviders(t *testing.T) {
	mod := testModule(t, "transform-provider-grandchild-inherit")
	concrete := func(a *NodeAbstractProvider) dag.Vertex { return a }

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		tf := &AttachResourceConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	{
		tf := TransformProviders([]string{"aws"}, concrete, mod)
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformModuleProviderGrandparentStr)
	if actual != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

// pass a specific provider into a module using it implicitly
func TestProviderConfigTransformer_implicitModule(t *testing.T) {
	mod := testModule(t, "transform-provider-implicit-module")
	concrete := func(a *NodeAbstractProvider) dag.Vertex { return a }

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		tf := &AttachResourceConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		tf := TransformProviders([]string{"aws"}, concrete, mod)
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(`module.mod.aws_instance.bar
  provider["registry.terraform.io/-/aws"].foo
provider["registry.terraform.io/-/aws"].foo`)
	if actual != expected {
		t.Fatalf("wrong result\n\nexpected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

// error out when a non-existent provider is named in a module providers map
func TestProviderConfigTransformer_invalidProvider(t *testing.T) {
	mod := testModule(t, "transform-provider-invalid")
	concrete := func(a *NodeAbstractProvider) dag.Vertex { return a }

	g := Graph{Path: addrs.RootModuleInstance}
	{
		tf := &ConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
	{
		tf := &AttachResourceConfigTransformer{Config: mod}
		if err := tf.Transform(&g); err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	tf := TransformProviders([]string{"aws"}, concrete, mod)
	err := tf.Transform(&g)
	if err == nil {
		t.Fatal("expected missing provider error")
	}
	if !strings.Contains(err.Error(), `provider["registry.terraform.io/-/aws"].foo`) {
		t.Fatalf("error should reference missing provider, got: %s", err)
	}
}

const testTransformProviderBasicStr = `
aws_instance.web
  provider["registry.terraform.io/-/aws"]
provider["registry.terraform.io/-/aws"]
`

const testTransformCloseProviderBasicStr = `
aws_instance.web
  provider["registry.terraform.io/-/aws"]
provider["registry.terraform.io/-/aws"]
provider["registry.terraform.io/-/aws"] (close)
  aws_instance.web
  provider["registry.terraform.io/-/aws"]
`

const testTransformMissingProviderBasicStr = `
aws_instance.web
  provider["registry.terraform.io/-/aws"]
foo_instance.web
  provider["registry.terraform.io/-/foo"]
provider["registry.terraform.io/-/aws"]
provider["registry.terraform.io/-/aws"] (close)
  aws_instance.web
  provider["registry.terraform.io/-/aws"]
provider["registry.terraform.io/-/foo"]
provider["registry.terraform.io/-/foo"] (close)
  foo_instance.web
  provider["registry.terraform.io/-/foo"]
`

const testTransformMissingGrandchildProviderStr = `
module.sub.module.subsub.bar_instance.two
  provider["registry.terraform.io/-/bar"]
module.sub.module.subsub.foo_instance.one
  module.sub.provider["registry.terraform.io/-/foo"]
module.sub.provider["registry.terraform.io/-/foo"]
provider["registry.terraform.io/-/bar"]
`

const testTransformMissingProviderModuleChildStr = `
module.moo.foo_instance.qux (import id "bar")
provider["registry.terraform.io/-/foo"]
`

const testTransformMissingProviderModuleGrandchildStr = `
module.a.module.b.foo_instance.qux (import id "bar")
provider["registry.terraform.io/-/foo"]
`

const testTransformParentProviderStr = `
module.moo.foo_instance.qux (import id "bar")
provider["registry.terraform.io/-/foo"]
`

const testTransformParentProviderModuleGrandchildStr = `
module.a.module.b.foo_instance.qux (import id "bar")
provider["registry.terraform.io/-/foo"]
`

const testTransformProviderModuleChildStr = `
module.moo.foo_instance.qux (import id "bar")
  provider["registry.terraform.io/-/foo"]
provider["registry.terraform.io/-/foo"]
`

const testTransformPruneProviderBasicStr = `
foo_instance.web
  provider["registry.terraform.io/-/foo"]
provider["registry.terraform.io/-/foo"]
provider["registry.terraform.io/-/foo"] (close)
  foo_instance.web
  provider["registry.terraform.io/-/foo"]
`

const testTransformDisableProviderBasicStr = `
module.child
  provider["registry.terraform.io/-/aws"] (disabled)
  var.foo
provider["registry.terraform.io/-/aws"] (close)
  module.child
  provider["registry.terraform.io/-/aws"] (disabled)
provider["registry.terraform.io/-/aws"] (disabled)
var.foo
`

const testTransformDisableProviderKeepStr = `
aws_instance.foo
  provider["registry.terraform.io/-/aws"]
module.child
  provider["registry.terraform.io/-/aws"]
  var.foo
provider["registry.terraform.io/-/aws"]
provider["registry.terraform.io/-/aws"] (close)
  aws_instance.foo
  module.child
  provider["registry.terraform.io/-/aws"]
var.foo
`

const testTransformModuleProviderConfigStr = `
module.child.aws_instance.thing
  provider["registry.terraform.io/-/aws"].foo
provider["registry.terraform.io/-/aws"].foo
`

const testTransformModuleProviderGrandparentStr = `
module.child.module.grandchild.aws_instance.baz
  provider["registry.terraform.io/-/aws"].foo
provider["registry.terraform.io/-/aws"].foo
`
