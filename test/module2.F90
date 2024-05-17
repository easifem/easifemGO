MODULE module2
USE Display_Method, ONLY: Display
IMPLICIT NONE
PRIVATE
PUBLIC :: baz

CONTAINS

FUNCTION baz() RESULT(ans)
  CHARACTER(len=10) :: ans
  ans = "Hello baz"
END FUNCTION baz

END MODULE module2
