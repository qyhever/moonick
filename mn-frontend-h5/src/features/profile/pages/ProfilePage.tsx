import type { FormEvent } from "react";
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import AvatarUploader from "../components/AvatarUploader";
import { getCurrentUserProfile, updateUserContact, updateUserProfile } from "../api";
import { getMyFavorites, getMyTrips } from "../../trips/api";
import { useAuthStore } from "../../../store/auth";

export default function ProfilePage() {
  const user = useAuthStore((state) => state.user);
  const userId = user?.id ?? null;
  const setUser = useAuthStore((state) => state.setUser);
  const logout = useAuthStore((state) => state.logout);
  const [avatarUrl, setAvatarUrl] = useState(user?.avatarUrl ?? "");
  const [nickname, setNickname] = useState(user?.nickname ?? "");
  const [defaultPhone, setDefaultPhone] = useState(user?.defaultPhone ?? "");
  const [defaultWechat, setDefaultWechat] = useState(user?.defaultWechat ?? "");
  const [tripCount, setTripCount] = useState(0);
  const [favoriteCount, setFavoriteCount] = useState(0);
  const [error, setError] = useState("");
  const [savingProfile, setSavingProfile] = useState(false);
  const [savingContact, setSavingContact] = useState(false);

  useEffect(() => {
    if (!userId) {
      return;
    }

    let active = true;

    async function loadProfile() {
      try {
        const [profile, myTrips, myFavorites] = await Promise.all([
          getCurrentUserProfile(),
          getMyTrips(),
          getMyFavorites(),
        ]);
        if (!active) {
          return;
        }

        setUser(profile);
        setAvatarUrl(profile.avatarUrl ?? "");
        setNickname(profile.nickname ?? "");
        setDefaultPhone(profile.defaultPhone ?? "");
        setDefaultWechat(profile.defaultWechat ?? "");
        setTripCount(myTrips.total);
        setFavoriteCount(myFavorites.total);
      } catch (loadError) {
        if (active) {
          setError(loadError instanceof Error ? loadError.message : "个人中心加载失败");
        }
      }
    }

    void loadProfile();

    return () => {
      active = false;
    };
  }, [setUser, userId]);

  async function handleProfileSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSavingProfile(true);
    setError("");

    try {
      await updateUserProfile(nickname);
      if (user) {
        setUser({ ...user, nickname });
      }
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "昵称更新失败");
    } finally {
      setSavingProfile(false);
    }
  }

  async function handleContactSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSavingContact(true);
    setError("");

    try {
      await updateUserContact(defaultWechat, defaultPhone);
      if (user) {
        setUser({ ...user, defaultWechat, defaultPhone });
      }
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "联系方式更新失败");
    } finally {
      setSavingContact(false);
    }
  }

  return (
    <main className="h5-shell">
      <section className="hero-card hero-card--compact">
        <p className="eyebrow">个人中心</p>
        <h1 className="hero-card__title">{user?.nickname || "旅途用户"}</h1>
        <p className="hero-card__subtitle">{user?.phone || "登录后可查看联系方式和偏好设置"}</p>
      </section>

      <section className="page-panel">
        <AvatarUploader
          initialUrl={avatarUrl}
          onUploaded={(nextUrl) => {
            setAvatarUrl(nextUrl);
            if (user) {
              setUser({ ...user, avatarUrl: nextUrl });
            }
          }}
        />

        <div className="profile-menu">
          <Link className="menu-link" to="/me/trips">
            <span>我的发布</span>
            <strong>{tripCount > 99 ? "99+" : tripCount}</strong>
          </Link>
          <Link className="menu-link" to="/me/favorites">
            <span>我的收藏</span>
            <strong>{favoriteCount > 99 ? "99+" : favoriteCount}</strong>
          </Link>
          <Link className="menu-link" to="/publish">
            再发一程
          </Link>
        </div>

        <form className="stack-form" onSubmit={handleProfileSubmit}>
          <label className="field-block">
            <span>昵称</span>
            <input onChange={(event) => setNickname(event.target.value)} value={nickname} />
          </label>
          <button className="primary-button" disabled={savingProfile} type="submit">
            保存昵称
          </button>
        </form>

        <form className="stack-form" onSubmit={handleContactSubmit}>
          <label className="field-block">
            <span>默认手机号</span>
            <input onChange={(event) => setDefaultPhone(event.target.value)} value={defaultPhone} />
          </label>
          <label className="field-block">
            <span>默认微信号</span>
            <input onChange={(event) => setDefaultWechat(event.target.value)} value={defaultWechat} />
          </label>
          <button className="primary-button" disabled={savingContact} type="submit">
            保存联系方式
          </button>
        </form>

        {error ? <p role="alert">{error}</p> : null}

        <button className="primary-button primary-button--ghost" onClick={logout} type="button">
          退出登录
        </button>
      </section>
    </main>
  );
}
