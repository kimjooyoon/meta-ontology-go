package semanticbinding

import (
	"fmt"
	"go/types"
)

func objectKey(object types.Object) string {
	key := object.Id()
	if function, ok := object.(*types.Func); ok {
		if signature, ok := function.Type().(*types.Signature); ok && signature.Recv() != nil {
			key += "|receiver=" + types.TypeString(signature.Recv().Type(), func(pkg *types.Package) string {
				if pkg == nil {
					return ""
				}
				return pkg.Path()
			})
		}
	}
	if key == "" {
		key = fmt.Sprintf("%T:%s", object, object.Name())
	}
	return key
}
