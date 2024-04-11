package main

// #cgo CFLAGS: -fPIC -DJULIA_INIT_DIR="C:/Users/gomezja/scoop/apps/julia/1.10.1/lib" -IC:\Users\gomezja\scoop\apps\julia\current\share\julia -I.
// #cgo LDFLAGS: -LC:\Users\gomezja\scoop\apps\julia\current\share\julia  -LC:/Users/gomezja/scoop/apps/julia/1.10.1/lib -Wl,-rpath,C:/Users/gomezja/scoop/apps/julia/1.10.1/lib -ljulia
// #include <julia.h>
import "C"

func main() {
	/* required: setup the Julia context */
	C.jl_init()

	/* run Julia commands */
	C.jl_eval_string(C.CString(`println(sqrt(2.0))`))

	/* strongly recommended: notify Julia that the
	   program is about to terminate. this allows
	   Julia time to cleanup pending write requests
	   and run all finalizers
	*/

	C.jl_atexit_hook(0)
}
