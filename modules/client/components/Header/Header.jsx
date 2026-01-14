"use client";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { useUserContext } from "../../context/userContext";
import HeaderAuthed from "../HeaderAuthed/HeaderAuthed";
import NotificationBell from "../NotificationBell/NotificationBell";
import { Menu, Button, useMantineColorScheme } from "@mantine/core";
import { ChevronDown, Settings, LogOut, BarChart2, Github, Linkedin, LogIn } from "lucide-react";

const Header = () => {
  const { userState } = useUserContext();
  const { user, isLoading } = userState;
  const pathname = usePathname();
  const { colorScheme } = useMantineColorScheme();
  const isDark = colorScheme === "dark";
  const isTenantSelectPage = pathname === "/tenant/select";
  const isLandingPage = pathname === "/";

  if (isLoading) return null;
  if (isLandingPage && !user) return null;

  return (
    <header className={`sticky top-0 z-40 border-b ${isDark ? "bg-zinc-800 border-zinc-700" : "bg-white border-zinc-200"}`}>
      <div
        className={`mx-auto max-w-6xl px-4 flex items-center justify-between ${user ? "py-0.5" : "py-4"}`}>
        <div className={`flex items-center ${user ? "gap-4" : "gap-6"}`}>
          <Link href="/" className="flex items-center gap-1.5">
            <Image
              src="/loading-icon.png"
              alt="PinaColada.co"
              width={user ? 24 : 48}
              height={user ? 24 : 48}
              priority
            />

            <span
              className={`font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500 ${user ? "text-base" : "text-3xl"}`}>
              PinaColada
            </span>
          </Link>
          <HeaderAuthed />
          {!user && (
            <div className="flex items-center gap-3">
              <Link
                href="https://github.com/hubenschmidt"
                target="_blank"
                className={`transition-colors duration-200 ${isDark ? "text-zinc-300 hover:text-white" : "text-zinc-600 hover:text-zinc-900"}`}>
                <Github size={20} />
              </Link>
              <Link
                href="https://www.linkedin.com/in/williamhubenschmidt"
                target="_blank"
                className={`transition-colors duration-200 ${isDark ? "text-zinc-300 hover:text-white" : "text-zinc-600 hover:text-zinc-900"}`}>
                <Linkedin size={20} />
              </Link>
            </div>
          )}
        </div>
        <div className="flex items-center gap-4">
          {/* Hide on mobile (portrait and landscape), show on tablets+ */}
          <nav className={`hidden sm:flex [@media(max-height:500px)]:!hidden items-center gap-6 text-sm font-semibold ${isDark ? "text-zinc-300" : "text-zinc-600"}`}>
            {!user && !isTenantSelectPage && (
              <>
                <Link href="/#services" className={`transition-colors duration-200 ${isDark ? "text-zinc-300 hover:text-white" : "text-zinc-600 hover:text-zinc-900"}`}>
                  Engineering
                </Link>
                <Link href="/#approach" className={`transition-colors duration-200 ${isDark ? "text-zinc-300 hover:text-white" : "text-zinc-600 hover:text-zinc-900"}`}>
                  Approach
                </Link>
                <Link href="/#contact" className={`transition-colors duration-200 ${isDark ? "text-zinc-300 hover:text-white" : "text-zinc-600 hover:text-zinc-900"}`}>
                  Free Consultation
                </Link>
              </>
            )}
          </nav>
          {/* Login/Account Menu */}
          {user ? (
            <div className="flex items-center gap-2">
              <NotificationBell />
              <Menu shadow="md" width={200}>
                <Menu.Target>
                  <Button
                    variant="subtle"
                    color="blue"
                    size="sm"
                    rightSection={<ChevronDown size={14} />}>
                    Account
                  </Button>
                </Menu.Target>
                <Menu.Dropdown>
                  <Menu.Item leftSection={<Settings size={16} />} component={Link} href="/settings">
                    Settings
                  </Menu.Item>
                  <Menu.Item leftSection={<BarChart2 size={16} />} component={Link} href="/usage">
                    Usage
                  </Menu.Item>
                  <Menu.Divider />
                  <Menu.Item
                    leftSection={<LogOut size={16} />}
                    component="a"
                    href="/auth/logout"
                    color="red">
                    Logout
                  </Menu.Item>
                </Menu.Dropdown>
              </Menu>
            </div>
          ) : (
            <Link href="/login" className={`transition-colors duration-200 ${isDark ? "text-zinc-300 hover:text-white" : "text-zinc-600 hover:text-zinc-900"}`}>
              <LogIn size={20} />
            </Link>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header;
