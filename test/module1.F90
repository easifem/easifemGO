MODULE module1
USE Display_Method, ONLY: Display
IMPLICIT NONE
PRIVATE
PUBLIC :: foo

CONTAINS

FUNCTION foo() RESULT(ans)
  CHARACTER(len=10) :: ans
  ans = "Hello foo"
END FUNCTION foo

END MODULE module1
