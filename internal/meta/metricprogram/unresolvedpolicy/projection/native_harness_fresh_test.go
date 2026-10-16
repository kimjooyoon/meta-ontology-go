package policyprojection

import "fmt"

func nativeDispositionHarness1866(flags bool) string {
	return fmt.Sprintf(`package unresolvedpolicy
import ("encoding/json"; "reflect"; "testing")
func TestNativeRecord1866(t *testing.T) {
 for _, row := range []struct {registered, recorded bool; resolution, reason, next string} {
  {true,true,"exact_operation","REGISTERED_SOURCE_OPERATION","CONTINUE_SOURCE_OPERATION"},
  {false,true,"invariant_only","OPERATION_NOT_IN_SOURCE_REGISTRY","DECLARE_SOURCE_OPERATION"},
  {true,false,"invariant_only","OPERATION_ORIGIN_REQUIRED","CAPTURE_OPERATION_ORIGIN"},
  {false,false,"invariant_only","OPERATION_ORIGIN_REQUIRED","CAPTURE_OPERATION_ORIGIN"},
 } {
  native := sourceOperationDisposition(row.registered,row.recorded)
  raw, err := json.Marshal(native); if err != nil {t.Fatal(err)}
  var got map[string]any; if err := json.Unmarshal(raw,&got); err != nil {t.Fatal(err)}
  want := map[string]any{"resolution":row.resolution,"reason":row.reason,"next_operation":row.next,
   "readiness_credit":%t,"repository_workspace_writes":%t,"promotion_authorized":%t}
  if !reflect.DeepEqual(got,want) {t.Fatalf("native field values: %%s",raw)}
  t.Logf("NATIVE_ROW1866 %%s",raw)
 }
}
`, flags, flags, flags)
}
