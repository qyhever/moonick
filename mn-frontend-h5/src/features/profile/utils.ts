export function maskEmail(email: string) {
  const trimmed = email.trim();
  if (!trimmed) {
    return "未设置邮箱";
  }

  const atIndex = trimmed.indexOf("@");
  if (atIndex <= 1) {
    return trimmed;
  }

  return `${trimmed.slice(0, 1)}***${trimmed.slice(atIndex - 1)}`;
}

export function maskPhone(phone: string) {
  if (phone.length < 7) {
    return phone || "未设置手机号";
  }

  return `${phone.slice(0, 3)}****${phone.slice(-4)}`;
}

export function getInitial(name: string) {
  return name.trim().slice(0, 1).toUpperCase() || "旅";
}
