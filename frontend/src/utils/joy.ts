export const moveFocusToJoyInput = (
  joyInput: React.RefObject<HTMLDivElement>,
): void => {
  const input = joyInput.current?.firstChild as HTMLInputElement | undefined
  input?.focus()
}
