import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { getCurrentUserProfile } from "../api";
import { getMyFavorites, getMyTrips } from "../../trips/api";
import { getInitial, maskPhone } from "../utils";
import { useAuthStore } from "../../../store/auth";

export default function ProfilePage() {
  const user = useAuthStore((state) => state.user);
  const userId = user?.id ?? null;
  const setUser = useAuthStore((state) => state.setUser);
  const [avatarUrl, setAvatarUrl] = useState(user?.avatarUrl ?? "");
  const [tripCount, setTripCount] = useState(0);
  const [favoriteCount, setFavoriteCount] = useState(0);
  const [error, setError] = useState("");

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

  const displayName = user?.nickname || "旅途用户";
  const displayPhone = maskPhone(user?.phone || user?.defaultPhone || "");

  return (
    <main className="h5-shell h5-shell--profile">
      <section className="profile-identity-card">
        <div className="profile-identity-card__top">
          <div className="profile-identity-card__avatar">
            {avatarUrl ? <img alt="当前头像" src={avatarUrl} /> : null}
            {!avatarUrl ? (
              <span className="profile-identity-card__avatar-fallback">{getInitial(displayName)}</span>
            ) : null}
          </div>
          <div className="profile-identity-card__meta">
            <p className="eyebrow">我的账户</p>
            <div className="profile-identity-card__title-row">
              <h1 className="hero-card__title">{displayName}</h1>
              <span className="profile-status-pill">已实名</span>
            </div>
            <p className="hero-card__subtitle">{displayPhone}</p>
          </div>
        </div>

        <div className="profile-stat-grid">
          <Link className="profile-stat-card" to="/me/trips">
            <span className="profile-stat-card__value">{tripCount > 99 ? "99+" : tripCount}</span>
            <span className="profile-stat-card__label">我的发布</span>
          </Link>
          <Link className="profile-stat-card" to="/me/favorites">
            <span className="profile-stat-card__value">{favoriteCount > 99 ? "99+" : favoriteCount}</span>
            <span className="profile-stat-card__label">我的收藏</span>
          </Link>
          <div className="profile-stat-card">
            <span className="profile-stat-card__value">98</span>
            <span className="profile-stat-card__label">信用评分</span>
          </div>
        </div>
      </section>

      <section className="page-panel profile-services-panel">
        <div className="section-header">
          <div>
            <h2 className="section-title">常用服务</h2>
            <p className="section-subtitle">高频入口集中收纳，首屏更干净</p>
          </div>
        </div>
        <div className="profile-service-grid">
          <button className="profile-service-item" type="button">
            <span className="profile-service-item__icon profile-service-item__icon--document">证</span>
            <strong>证件信息</strong>
            <span>实名认证资料</span>
          </button>
          <button className="profile-service-item" type="button">
            <span className="profile-service-item__icon profile-service-item__icon--wallet">¥</span>
            <strong>钱包余额</strong>
            <span>收支与明细</span>
          </button>
          <button className="profile-service-item" type="button">
            <span className="profile-service-item__icon profile-service-item__icon--passenger">人</span>
            <strong>常用乘客</strong>
            <span>出行档案管理</span>
          </button>
          <button className="profile-service-item" type="button">
            <span className="profile-service-item__icon profile-service-item__icon--help">?</span>
            <strong>客服帮助</strong>
            <span>平台与出行支持</span>
          </button>
        </div>
      </section>

      <section className="page-panel profile-settings-panel">
        <div className="section-header">
          <div>
            <h2 className="section-title">账户与安全</h2>
            <p className="section-subtitle">账户设置已拆为独立页面，这里只保留进入入口</p>
          </div>
        </div>
        <Link className="profile-setting-entry" to="/me/settings">
          <div>
            <p className="profile-setting-item__label">账户设置</p>
            <p className="profile-setting-item__value">管理昵称、手机号、登录安全与退出登录</p>
          </div>
          <span className="profile-setting-entry__action">去查看</span>
        </Link>
      </section>

      {error ? <p role="alert">{error}</p> : null}
    </main>
  );
}
