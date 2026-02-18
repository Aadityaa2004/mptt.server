"use client";

import Link from "next/link";
import Image from "next/image";
import { LogOut, User } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/AuthContext";

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();

  const handleLogout = async () => {
    await logout();
  };

  return (
    <nav className="fixed left-4 right-4 sm:left-6 sm:right-6 z-50 rounded-2xl border-white/10 bg-black/60 backdrop-blur-md shadow-lg">
      <div className="mx-auto px-4 sm:px-6 lg:px-8">
        <div className="relative flex h-16 items-center">
          {/* Logo */}
          {isAuthenticated ? (
            <div className="flex items-center gap-2 flex-shrink-0">
              <Image
                src="/maple_sense_logo.png"
                alt="MapleSense Logo"
                width={32}
                height={32}
                className="object-contain"
              />
              <span className="text-lg font-light text-white hidden sm:block">MapleSense</span>
            </div>
          ) : (
            <Link href="/" className="flex items-center gap-2 flex-shrink-0">
              <Image
                src="/maple_sense_logo.png"
                alt="MapleSense Logo"
                width={32}
                height={32}
                className="object-contain"
              />
              <span className="text-lg font-light text-white hidden sm:block">MapleSense</span>
            </Link>
          )}

          {/* Navigation Links - Absolutely centered, Hidden on mobile, visible on desktop */}
          <div className="hidden md:flex items-center gap-6 absolute left-1/2 -translate-x-1/2">
            {isAuthenticated ? (
              <>
                <Link
                  href={user?.role === "admin" ? "/admin/dashboard" : "/user/dashboard"}
                  className="text-sm font-light text-white/90 hover:text-white transition-colors"
                >
                  Dashboard
                </Link>
                {user?.role === "admin" && (
                  <>
                    <Link
                      href="/admin/sensors"
                      className="text-sm font-light text-white/90 hover:text-white transition-colors"
                    >
                      Sensors
                    </Link>
                    <Link
                      href="/admin/settings"
                      className="text-sm font-light text-white/90 hover:text-white transition-colors"
                    >
                      Settings
                    </Link>
                  </>
                )}
                {user?.role === "user" && (
                  <>
                    <Link
                      href="/user/forecast"
                      className="text-sm font-light text-white/90 hover:text-white transition-colors"
                    >
                      Forecast
                    </Link>
                    <Link
                      href="/user/sensors"
                      className="text-sm font-light text-white/90 hover:text-white transition-colors"
                    >
                      My Sensors
                    </Link>
                    <Link
                      href="/user/settings"
                      className="text-sm font-light text-white/90 hover:text-white transition-colors"
                    >
                      Settings
                    </Link>
                  </>
                )}
              </>
            ) : (
              <>
                <Link
                  href="/"
                  className="text-sm font-light text-white/90 hover:text-white transition-colors"
                >
                  Home
                </Link>
                <Link
                  href="/about-us"
                  className="text-sm font-light text-white/90 hover:text-white transition-colors"
                >
                  About Us
                </Link>
                <Link
                  href="/contact-us"
                  className="text-sm font-light text-white/90 hover:text-white transition-colors"
                >
                  Contact Us
                </Link>
                <Link
                  href="/products"
                  className="text-sm font-light text-white/90 hover:text-white transition-colors"
                >
                  Products
                </Link>
              </>
            )}
          </div>

          {/* Right Side Actions */}
          <div className="flex items-center gap-3 flex-shrink-0 ml-auto">
            {isAuthenticated ? (
              <>
                {/* Map View Button - Only for authenticated users */}
                {/* <Button
                  variant="ghost"
                  size="sm"
                  className="hidden sm:flex items-center gap-1.5 text-white/90 hover:text-white hover:bg-white/10 border border-white/20 rounded-full px-4 h-8"
                >
                  <MapPin className="h-3.5 w-3.5" />
                  <span className="text-xs font-light">FAQ</span>
                  <ExternalLink className="h-3 w-3" />
                </Button> */}

                {/* Search Icon */}
                {/* <Button
                  variant="ghost"
                  size="icon"
                  className="text-white/90 hover:text-white hover:bg-white/10 h-8 w-8"
                >
                  <Search className="h-4 w-4" />
                </Button> */}

                {/* User Info */}
                <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-gray-300 border rounded-md border-white/20">
                  <User className="h-3.5 w-3.5 text-black" />
                  <span className="text-xs font-light text-black">{user?.username}</span>
                  {user?.role === "admin" && (
                    <span className="text-xs font-light text-black">(Admin)</span>
                  )}
                </div>

                {/* Logout Button */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleLogout}
                  className="text-white bg-orange-500/85 border-2 border-white hover:bg-orange-500/70 hover:border-white/30 text-xs font-light h-8 px-4  flex items-center gap-1.5"
                >
                  <LogOut className="h-3.5 w-3.5" />
                  <span className="hidden sm:inline">Logout</span>
                </Button>
              </>
            ) : (
              <>
                {/* Map View Button */}
                {/* <Button
                  variant="ghost"
                  size="sm"
                  className="hidden sm:flex items-center gap-1.5 text-white/90 hover:text-white hover:bg-white/10 border-2 border-white rounded-full px-4 h-8"
                >
                  
                  <span className="text-xs font-light">FAQ</span>
                  <ExternalLink className="h-3 w-3" />
                </Button> */}

                {/* Search Icon */}
                {/* <Button
                  variant="ghost"
                  size="icon"
                  className="text-white/90 hover:text-white hover:bg-white/10 h-8 w-8"
                >
                  <Search className="h-4 w-4" />
                </Button> */}

                {/* Create Account Button */}
                <Link href="/register">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="hover:bg-white/90 hover:text-black text-xs font-light h-8 px-4 bg-white text-black"
                  >
                    Create Account
                  </Button>
                </Link>

                {/* Login Button */}
                <Link href="/login">
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-white border-2 border-white hover:bg-orange-500/70 hover:border-white/70 text-xs font-light h-8 px-4 bg-orange-500/85"
                  >
                    Login
                  </Button>
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}

