"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { LogOut, User, Menu, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/AuthContext";

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const handleLogout = async () => {
    await logout();
    setMobileMenuOpen(false);
  };

  const NavLinks = () => (
    <>
      {isAuthenticated ? (
        <>
          <Link
            href={user?.role === "admin" ? "/admin/overview" : "/user/dashboard"}
            className="text-foreground/90 hover:text-foreground transition-colors"
            onClick={() => setMobileMenuOpen(false)}
          >
            {user?.role === "admin" ? "Overview" : "Dashboard"}
          </Link>
          {user?.role === "admin" && (
            <>
              <Link href="/admin/users" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Users</Link>
              <Link href="/admin/pis" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>PIs</Link>
              <Link href="/admin/devices" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Devices</Link>
              <Link href="/admin/readings" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Readings</Link>
              <Link href="/admin/settings" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Settings</Link>
            </>
          )}
          {user?.role === "user" && (
            <>
              <Link href="/user/forecast" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Forecast</Link>
              <Link href="/user/sensors" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>My Sensors</Link>
              <Link href="/user/settings" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Settings</Link>
            </>
          )}
        </>
      ) : (
        <>
          <Link href="/" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Home</Link>
          <Link href="/about-us" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>About Us</Link>
          <Link href="/contact-us" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Contact Us</Link>
          <Link href="/products" className="text-foreground/90 hover:text-foreground transition-colors" onClick={() => setMobileMenuOpen(false)}>Products</Link>
        </>
      )}
    </>
  );

  return (
    <>
      <nav className="fixed left-2 right-2 sm:left-4 sm:right-4 md:left-6 md:right-6 top-2 sm:top-4 z-50 rounded-xl md:rounded-2xl border-border bg-background/95 backdrop-blur-md shadow-lg dark:bg-black/80 dark:border-white/10">
        <div className="mx-auto px-3 sm:px-6 lg:px-8">
          <div className="relative flex h-14 sm:h-16 items-center">
            {/* Mobile menu button */}
            <button
              type="button"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="md:hidden -ml-2 p-2 rounded-lg text-foreground/90 hover:bg-white/10 mr-2"
              aria-label="Toggle menu"
            >
              {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </button>

            {/* Logo */}
            {isAuthenticated ? (
              <div className="flex items-center gap-2 flex-shrink-0">
                <Image src="/maple_sense_logo.png" alt="MapleSense" width={28} height={28} className="object-contain sm:w-8 sm:h-8" />
                <span className="text-base sm:text-lg font-light text-foreground hidden sm:block">MapleSense</span>
              </div>
            ) : (
              <Link href="/" className="flex items-center gap-2 flex-shrink-0">
                <Image src="/maple_sense_logo.png" alt="MapleSense" width={28} height={28} className="object-contain sm:w-8 sm:h-8" />
                <span className="text-base sm:text-lg font-light text-foreground hidden sm:block">MapleSense</span>
              </Link>
            )}

            {/* Desktop Navigation - centered */}
            <div className="hidden md:flex items-center gap-6 absolute left-1/2 -translate-x-1/2 text-sm font-light">
              <NavLinks />
            </div>

          {/* Right Side Actions */}
          <div className="flex items-center gap-2 sm:gap-3 flex-shrink-0 ml-auto">
            {isAuthenticated ? (
              <>
                <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-gray-300 border rounded-md border-white/20">
                  <User className="h-3.5 w-3.5 text-black" />
                  <span className="text-xs font-light text-black">{user?.username}</span>
                  {user?.role === "admin" && (
                    <span className="text-xs font-light text-black">(Admin)</span>
                  )}
                </div>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleLogout}
                  className="text-white bg-orange-500/85 border-2 border-white hover:bg-orange-500/70 hover:border-white/30 text-xs font-light h-8 sm:h-8 px-3 sm:px-4 flex items-center gap-1.5"
                >
                  <LogOut className="h-3.5 w-3.5" />
                  <span className="hidden sm:inline">Logout</span>
                </Button>
              </>
            ) : (
              <>
                <Link href="/register" className="hidden sm:block">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="hover:bg-white/90 hover:text-black text-xs font-light h-8 px-3 sm:px-4 bg-white text-black"
                  >
                    Create Account
                  </Button>
                </Link>
                <Link href="/login">
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-white border-2 border-white hover:bg-orange-500/70 hover:border-white/70 text-xs font-light h-8 px-3 sm:px-4 bg-orange-500/85"
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

      {/* Mobile menu overlay */}
      {mobileMenuOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm md:hidden"
          onClick={() => setMobileMenuOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* Mobile menu panel */}
      <div
        className={`fixed top-0 right-0 z-50 h-full w-[280px] max-w-[85vw] bg-background/98 backdrop-blur-xl border-l border-white/10 shadow-xl transform transition-transform duration-200 ease-out md:hidden ${
          mobileMenuOpen ? "translate-x-0" : "translate-x-full"
        }`}
      >
        <div className="flex flex-col h-full pt-24 px-4 pb-6">
          <div className="flex flex-col [&_a]:block [&_a]:py-3 [&_a]:px-3 [&_a]:rounded-lg [&_a]:text-base [&_a]:font-light [&_a]:hover:bg-white/5">
            <NavLinks />
          </div>
          {isAuthenticated && (
            <div className="mt-auto pt-4 border-t border-white/10">
              <div className="flex items-center gap-2 px-3 py-2 text-sm text-foreground/80 mb-3">
                <User className="h-4 w-4" />
                <span>{user?.username}</span>
                {user?.role === "admin" && <span className="text-xs">(Admin)</span>}
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={handleLogout}
                className="w-full justify-center text-white bg-orange-500/85 border-white/20"
              >
                <LogOut className="h-4 w-4 mr-2" />
                Logout
              </Button>
            </div>
          )}
        </div>
      </div>
    </>
  );
}

