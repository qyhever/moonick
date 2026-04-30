export function maskPhone(phone: string) {
  if (phone.length < 7) {
    return phone || "未设置手机号";
  }

  return `${phone.slice(0, 3)}****${phone.slice(-4)}`;
}

export function getInitial(name: string) {
  return name.trim().slice(0, 1).toUpperCase() || "旅";
}
