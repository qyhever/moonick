import type { ReactNode } from "react";
import { Image } from "antd-mobile";

type UserAvatarProps = {
  src?: string;
  alt: string;
  fallback: ReactNode;
  defaultSrc?: string;
  className?: string;
  fallbackClassName?: string;
};

export default function UserAvatar({
  src = "",
  alt,
  fallback,
  defaultSrc = "",
  className,
  fallbackClassName,
}: UserAvatarProps) {
  const fallbackNode =
    typeof fallback === "string" || typeof fallback === "number" ? (
      <span className={fallbackClassName}>{fallback}</span>
    ) : (
      fallback
    );

  const resolvedSrc = src || defaultSrc;

  if (!resolvedSrc) {
    return fallbackNode;
  }

  return (
    <Image
      alt={alt}
      className={className}
      fallback={fallbackNode}
      fit="cover"
      placeholder={fallbackNode}
      src={resolvedSrc}
    />
  );
}
