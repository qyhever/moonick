export function formatDateTime(value?: string | null) {
  if (!value) {
    return "-";
  }

  const matched = value.match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2}:\d{2}:\d{2})/);
  if (matched) {
    return `${matched[1]} ${matched[2]}`;
  }

  return value;
}
