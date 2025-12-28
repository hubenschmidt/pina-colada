"use client";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { useNavContext } from "../../context/navContext";
import { useUserContext } from "../../context/userContext";
import HeaderAuthed from "../HeaderAuthed/HeaderAuthed";
import NotificationBell from "../NotificationBell/NotificationBell";
import { Menu, Button } from "@mantine/core";
import { ChevronDown, Settings, LogOut, BarChart2 } from "lucide-react";

const Header = () => {
  const { dispatchNav } = useNavContext();
  const { userState } = useUserContext();
  const { user, isLoading } = userState;
  const pathname = usePathname();
  const isTenantSelectPage = pathname === "/tenant/select";

  if (isLoading) return null;

  return (
    <header className="sticky top-0 z-40 bg-blue-50 dark:bg-zinc-900">
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
        </div>
        <div className="flex items-center gap-4">
          {/* Hide on mobile (portrait and landscape), show on tablets+ */}
          <nav className="hidden sm:flex [@media(max-height:500px)]:!hidden items-center gap-6 text-sm text-zinc-600 font-semibold">
            {!user && !isTenantSelectPage && (
              <>
                <Link
                  href="/#agent"
                  className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 hover:text-blue-500"
                  onClick={() => dispatchNav({ type: "SET_AGENT_OPEN", payload: true })}>
                  Chat with us
                </Link>
                <Link
                  href="/about"
                  className="text-blue-700 hover:text-blue-500"
                  onClick={() => dispatchNav({ type: "SET_AGENT_OPEN", payload: false })}>
                  About
                </Link>
                <Link
                  href="/#services"
                  className="text-blue-700 hover:text-blue-500"
                  onClick={() => dispatchNav({ type: "SET_AGENT_OPEN", payload: false })}>
                  Software Development
                </Link>
                <Link
                  href="/#ai"
                  className="text-blue-700 hover:text-blue-500"
                  onClick={() => dispatchNav({ type: "SET_AGENT_OPEN", payload: false })}>
                  AI
                </Link>
                <Link
                  href="/#approach"
                  className="text-blue-700 hover:text-blue-500"
                  onClick={() => dispatchNav({ type: "SET_AGENT_OPEN", payload: false })}>
                  Approach
                </Link>
                <Link
                  href="/#portfolio"
                  className="text-blue-700 hover:text-blue-500"
                  onClick={() => dispatchNav({ type: "SET_AGENT_OPEN", payload: false })}>
                  Portfolio
                </Link>
                <Link
                  href="/#contact"
                  className="text-blue-700 hover:text-blue-500"
                  onClick={() => dispatchNav({ type: "SET_AGENT_OPEN", payload: false })}>
                  Contact
                </Link>
                <Link
                  href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
                  className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 hover:text-blue-500"
                  onClick={() => dispatchNav({ type: "SET_AGENT_OPEN", payload: false })}>
                  Start a project
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
            <Link href="/login" className="text-blue-700 hover:text-blue-500 text-sm font-semibold">
              Login
            </Link>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header;
