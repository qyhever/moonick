import type { FormEvent } from "react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getCurrentUserProfile, updateUserContact, updateUserProfile } from "../api";
import AvatarUploader from "../components/AvatarUploader";
import UserAvatar from "../components/UserAvatar";
import { AVATAR_DEFAULT_SRC, AVATAR_FALLBACK_SRC } from "../components/avatarAssets";
import { maskEmail } from "../utils";
import { useAuthStore } from "../../../store/auth";

export default function AccountSettingsPage() {
  const accessToken = useAuthStore((state) => state.accessToken);
  const user = useAuthStore((state) => state.user);
  const setUser = useAuthStore((state) => state.setUser);
  const logout = useAuthStore((state) => state.logout);
  const [avatarUrl, setAvatarUrl] = useState(user?.avatarUrl ?? "");
  const [nickname, setNickname] = useState(user?.nickname ?? "");
  const [defaultPhone, setDefaultPhone] = useState(user?.defaultPhone ?? "");
  const [defaultWechat, setDefaultWechat] = useState(user?.defaultWechat ?? "");
  const [error, setError] = useState("");
  const [savingProfile, setSavingProfile] = useState(false);
  const [editingField, setEditingField] = useState<"nickname" | "contact" | null>(null);

  useEffect(() => {
    if (!accessToken) {
      return;
    }

    let active = true;

    async function loadProfile() {
      try {
        const profile = await getCurrentUserProfile();
        if (!active) {
          return;
        }

        setUser(profile);
        setAvatarUrl(profile.avatarUrl ?? "");
        setNickname(profile.nickname ?? "");
        setDefaultPhone(profile.defaultPhone ?? "");
        setDefaultWechat(profile.defaultWechat ?? "");
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : "账户设置加载失败");
        }
      }
    }

    void loadProfile();

    return () => {
      active = false;
    };
  }, [accessToken, setUser]);

  const displayName = nickname || user?.nickname || "旅途用户";
  const displayEmail = maskEmail(user?.email || "");
  const hasNicknameChange = nickname !== (user?.nickname ?? "");
  const hasContactChange =
    defaultPhone !== (user?.defaultPhone ?? "") || defaultWechat !== (user?.defaultWechat ?? "");
  const hasChanges = hasNicknameChange || hasContactChange;

  async function handleProfileSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSavingProfile(true);
    setError("");

    try {
      if (hasNicknameChange) {
        await updateUserProfile(nickname);
      }
      if (hasContactChange) {
        await updateUserContact(defaultWechat, defaultPhone);
      }
      if (user) {
        setUser({
          ...user,
          nickname,
          defaultWechat,
          defaultPhone,
          avatarUrl,
        });
      }
      setEditingField(null);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "资料更新失败");
    } finally {
      setSavingProfile(false);
    }
  }

  return (
    <main className="h5-shell h5-shell--profile">
      <section className="profile-identity-card">
        <div className="profile-identity-card__top">
          <div className="profile-identity-card__avatar">
            <UserAvatar
              alt="当前头像"
              className="profile-identity-card__avatar-image"
              defaultSrc={AVATAR_DEFAULT_SRC}
              fallback={<img alt="头像加载失败" className="profile-identity-card__avatar-fallback" src={AVATAR_FALLBACK_SRC} />}
              fallbackClassName="profile-identity-card__avatar-fallback"
              src={avatarUrl}
            />
          </div>
          <div className="profile-identity-card__meta">
            <p className="eyebrow">资料与安全</p>
            <div className="profile-identity-card__title-row">
              <h1 className="hero-card__title">账户设置</h1>
            </div>
            <p className="hero-card__subtitle">{displayName} · {displayEmail}</p>
          </div>
        </div>
      </section>

      <section className="page-panel profile-settings-panel">
        <div className="page-header">
          <div>
            <p className="eyebrow">账户设置</p>
            <h1>资料与安全管理</h1>
          </div>
        </div>

        <div className="profile-avatar-editor">
          <AvatarUploader
            initialUrl={avatarUrl}
            onUploaded={(nextUrl) => {
              setAvatarUrl(nextUrl);
              if (user) {
                setUser({ ...user, avatarUrl: nextUrl });
              }
            }}
          />
        </div>

        <form className="profile-settings-form" onSubmit={handleProfileSubmit}>
          <section className="profile-settings-group">
            <div className="section-header">
              <div>
                <h2 className="section-title">基本资料</h2>
                <p className="section-subtitle">优先展示用户最常修改的信息</p>
              </div>
            </div>

            <div className="profile-setting-list">
              <div className="profile-setting-item">
                <div>
                  <p className="profile-setting-item__label">昵称</p>
                  <p className="profile-setting-item__value">{displayName}</p>
                </div>
                <button
                  className="secondary-link secondary-link--button"
                  onClick={() => setEditingField(editingField === "nickname" ? null : "nickname")}
                  type="button"
                >
                  {editingField === "nickname" ? "收起" : "修改"}
                </button>
              </div>

              {editingField === "nickname" ? (
                <label className="field-block">
                  <span>修改昵称</span>
                  <input onChange={(event) => setNickname(event.target.value)} value={nickname} />
                </label>
              ) : null}

              <div className="profile-setting-item">
                <div>
                  <p className="profile-setting-item__label">登录邮箱</p>
                  <p className="profile-setting-item__value">{user?.email || "未设置邮箱"}</p>
                </div>
              </div>

              <div className="profile-setting-item">
                <div>
                  <p className="profile-setting-item__label">默认手机号</p>
                  <p className="profile-setting-item__value">{defaultPhone || "未设置默认联系方式"}</p>
                </div>
                <button
                  className="secondary-link secondary-link--button"
                  onClick={() => setEditingField(editingField === "contact" ? null : "contact")}
                  type="button"
                >
                  {editingField === "contact" ? "收起" : "去管理"}
                </button>
              </div>

              {editingField === "contact" ? (
                <div className="profile-inline-fields">
                  <label className="field-block">
                    <span>默认手机号</span>
                    <input onChange={(event) => setDefaultPhone(event.target.value)} value={defaultPhone} />
                  </label>
                  <label className="field-block">
                    <span>默认微信号</span>
                    <input onChange={(event) => setDefaultWechat(event.target.value)} value={defaultWechat} />
                  </label>
                </div>
              ) : null}
            </div>
          </section>

          <section className="profile-settings-group">
            <div className="section-header">
              <div>
                <h2 className="section-title">安全管理</h2>
                <p className="section-subtitle">为后续密码、设备与风控能力预留结构</p>
              </div>
            </div>

            <div className="profile-setting-list">
              <Link className="profile-setting-entry" to="/me/settings/password-reset">
                <div>
                  <p className="profile-setting-item__label">重置密码</p>
                  <p className="profile-setting-item__value">定期更新登录密码，保护账户登录安全</p>
                </div>
                <span className="profile-setting-entry__action">去修改</span>
              </Link>
            </div>
          </section>

          {error ? <p role="alert">{error}</p> : null}

          <div className="profile-action-row">
            <button className="primary-button" disabled={savingProfile || !hasChanges} type="submit">
              {savingProfile ? "保存中..." : "保存修改"}
            </button>
            <button className="primary-button primary-button--ghost primary-button--danger" onClick={logout} type="button">
              退出登录
            </button>
          </div>
        </form>
      </section>
    </main>
  );
}
