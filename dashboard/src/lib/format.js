export function formatDateTime(value) {
  if (!value || String(value).startsWith("0001-01-01")) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }

  return date.toLocaleString("tr-TR");
}

export function getHealthStateLabel(source) {
  if (isFutureDate(source.disabledUntil)) {
    return "disabled";
  }

  if ((source.consecutiveFails ?? 0) > 0) {
    return "degraded";
  }

  return "healthy";
}

function isFutureDate(value) {
  if (!value || String(value).startsWith("0001-01-01")) {
    return false;
  }

  const date = new Date(value);
  return !Number.isNaN(date.getTime()) && date.getTime() > Date.now();
}
