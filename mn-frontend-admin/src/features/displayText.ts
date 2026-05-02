const tripTypeTextMap: Record<string, string> = {
  driver_post: "车找人",
  passenger_post: "人找车",
};

const tripStatusTextMap: Record<string, string> = {
  active: "可约",
  full: "已满",
  closed: "已关闭",
  expired: "已过期",
};

const userStatusTextMap: Record<string, string> = {
  active: "正常",
  disabled: "已禁用",
};

export function getTripTypeText(value: string) {
  return tripTypeTextMap[value] ?? value;
}

export function getTripStatusText(value: string) {
  return tripStatusTextMap[value] ?? value;
}

export function getUserStatusText(value: string) {
  return userStatusTextMap[value] ?? value;
}
