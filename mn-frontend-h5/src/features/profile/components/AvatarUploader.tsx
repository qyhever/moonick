import { useEffect, useState } from "react";

import { uploadAvatar } from "../../trips/api";

type AvatarUploaderProps = {
  initialUrl?: string;
  onUploaded?: (url: string) => void;
};

export default function AvatarUploader({
  initialUrl = "",
  onUploaded,
}: AvatarUploaderProps) {
  const [avatarUrl, setAvatarUrl] = useState(initialUrl);
  const [error, setError] = useState("");
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    setAvatarUrl(initialUrl);
  }, [initialUrl]);

  async function onUpload(file: File) {
    const previous = avatarUrl;
    const localPreview = URL.createObjectURL(file);
    setAvatarUrl(localPreview);
    setUploading(true);
    setError("");

    try {
      const nextUrl = await uploadAvatar(file);
      setAvatarUrl(nextUrl);
      onUploaded?.(nextUrl);
    } catch {
      setAvatarUrl(previous);
      onUploaded?.(previous);
      setError("服务器异常，请稍后再试");
    } finally {
      if (typeof URL.revokeObjectURL === "function") {
        URL.revokeObjectURL(localPreview);
      }
      setUploading(false);
    }
  }

  return (
    <div className="avatar-uploader">
      {avatarUrl ? (
        <img alt="当前头像" className="avatar-uploader__image" src={avatarUrl} />
      ) : (
        <div className="avatar-uploader__placeholder" aria-hidden="true">
          头像
        </div>
      )}

      <div className="avatar-uploader__info">
        <strong>头像与个人资料</strong>
        <p className="subtle-text">失败时会自动回退到上一次成功头像。</p>
      </div>

      <label className="primary-button" htmlFor="avatar-upload-input">
        上传头像
      </label>
      <input
        accept=".jpg,.jpeg,.png,.webp,.heic,image/jpeg,image/png,image/webp,image/heic"
        id="avatar-upload-input"
        className="sr-only"
        disabled={uploading}
        onChange={(event) => {
          const file = event.target.files?.[0];
          if (file) {
            if (file.size > 10 * 1024 * 1024) {
              setError("头像大小不能超过 10 MB");
              event.target.value = "";
              return;
            }
            void onUpload(file);
          }
          event.target.value = "";
        }}
        type="file"
      />

      {error ? <p role="alert">{error}</p> : null}
    </div>
  );
}
