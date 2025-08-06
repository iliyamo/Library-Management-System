-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Aug 03, 2025 at 03:32 PM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.2.12

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `digital_library`
--

-- --------------------------------------------------------

--
-- Table structure for table `authors`
--

CREATE TABLE `authors` (
  `id` int(11) NOT NULL,
  `name` varchar(100) NOT NULL,
  `biography` text DEFAULT NULL,
  `birth_date` date DEFAULT NULL,
  `created_at` datetime DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `authors`
--

INSERT INTO `authors` (`id`, `name`, `biography`, `birth_date`, `created_at`) VALUES
(7, 'ایلیا', '17 year\'s old writer', '2007-12-27', '2025-07-02 11:34:08'),
(8, 'ایلیا', '17 year\'s old writer', '2007-12-27', '2025-07-02 11:34:09'),
(9, 'ایلیا', '17 year\'s old writer', '2007-12-27', '2025-07-02 11:36:46'),
(10, 'ایلیا', '17 year\'s old writer', '2007-12-27', '2025-07-02 11:36:47'),
(11, 'ایلیا', '17 year\'s old writer', '2007-12-27', '2025-07-02 11:44:59'),
(12, 'احمد', 'شاعر قرن 20', '1985-05-20', '2025-07-02 11:45:00'),
(13, 'iliya', '17 year\'s old writer', '2007-12-27', '2025-07-13 08:46:43'),
(14, 'iliyamo', '17 year\'s old writer', '2007-12-27', '2025-07-13 08:48:13'),
(16, 'William Shakespeare', 'English playwright, poet, and actor. Widely regarded as the greatest writer in the English language.', '1564-04-23', '2025-07-20 15:31:50'),
(17, 'Jane Austen', 'Known for her six major novels, which interpret, critique and comment upon the British landed gentry.', '1775-12-16', '2025-07-20 15:31:50'),
(18, 'Mark Twain', 'American writer, humorist, entrepreneur, publisher, and lecturer.', '1835-11-30', '2025-07-20 15:31:50'),
(19, 'Virginia Woolf', 'English writer, considered one of the most important modernist 20th-century authors.', '1882-01-25', '2025-07-20 15:31:50'),
(20, 'George Orwell', 'English novelist, essayist, journalist and critic. Famous for Animal Farm and 1984.', '1903-06-25', '2025-07-20 15:31:50'),
(21, 'Jane', '36 year\'s old writer', '1990-12-27', '2025-07-20 12:22:10');

-- --------------------------------------------------------

--
-- Table structure for table `books`
--

CREATE TABLE `books` (
  `id` int(11) NOT NULL,
  `title` varchar(150) NOT NULL,
  `isbn` varchar(20) NOT NULL,
  `author_id` int(11) NOT NULL,
  `category_id` int(11) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `published_year` int(11) DEFAULT NULL,
  `total_copies` int(11) NOT NULL DEFAULT 1,
  `available_copies` int(11) NOT NULL DEFAULT 1,
  `created_at` datetime DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `books`
--

INSERT INTO `books` (`id`, `title`, `isbn`, `author_id`, `category_id`, `description`, `published_year`, `total_copies`, `available_copies`, `created_at`) VALUES
(14, 'کتاب تستی', '978-964-312-452-3', 8, NULL, 'این یک کتاب تستی است', 2023, 5, 5, '2025-07-02 12:27:27'),
(15, 'کتاب تستی', '978-964-912-452-3', 8, NULL, 'این یک کتاب تستی است', 2023, 5, 5, '2025-07-02 12:34:33'),
(16, 'book', '978-962-912-452-3', 8, NULL, 'این یک کتاب تستی است', 2023, 5, 5, '2025-07-13 10:40:26'),
(17, 'go', '978-962-912-442-3', 8, NULL, 'این یک کتاب تستی است', 2023, 5, 4, '2025-07-13 10:44:20'),
(19, 'python by x', '278-462-912-442-3', 8, NULL, 'این یک کتاب تستی است', 2023, 5, 5, '2025-07-13 10:46:03'),
(20, 'python by y', '278-462-312-442-3', 8, NULL, 'این یک کتاب تستی است', 2023, 5, 5, '2025-07-13 10:46:09'),
(21, 'python by y', '278-362-312-442-3', 8, NULL, 'این یک کتاب تستی است', 2023, 5, 5, '2025-08-03 13:26:19');

-- --------------------------------------------------------

--
-- Table structure for table `book_reviews`
--

CREATE TABLE `book_reviews` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `book_id` int(11) NOT NULL,
  `rating` tinyint(4) NOT NULL CHECK (`rating` between 1 and 5),
  `comment` text DEFAULT NULL,
  `created_at` datetime DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `categories`
--

CREATE TABLE `categories` (
  `id` int(11) NOT NULL,
  `name` varchar(50) NOT NULL,
  `description` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `loans`
--

CREATE TABLE `loans` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `book_id` int(11) NOT NULL,
  `loan_date` date NOT NULL,
  `due_date` date NOT NULL,
  `return_date` date DEFAULT NULL,
  `status` enum('borrowed','returned','late') DEFAULT 'borrowed'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `loans`
--

INSERT INTO `loans` (`id`, `user_id`, `book_id`, `loan_date`, `due_date`, `return_date`, `status`) VALUES
(1, 6, 14, '2025-08-03', '2025-08-10', '2025-08-03', 'returned'),
(2, 6, 14, '2025-08-03', '2025-08-10', '2025-08-03', 'returned'),
(3, 6, 15, '2025-08-03', '2025-08-10', '2025-08-03', 'returned'),
(4, 6, 15, '2025-08-03', '2025-08-10', '2025-08-03', 'returned'),
(5, 6, 16, '2025-08-03', '2025-08-10', '2025-08-03', 'returned'),
(6, 6, 17, '2025-08-03', '2025-08-10', '2025-08-03', 'returned'),
(7, 6, 17, '2025-08-03', '2025-08-10', NULL, 'borrowed');

-- --------------------------------------------------------

--
-- Table structure for table `refresh_tokens`
--

CREATE TABLE `refresh_tokens` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `token` text NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` datetime DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `refresh_tokens`
--

INSERT INTO `refresh_tokens` (`id`, `user_id`, `token`, `expires_at`, `created_at`) VALUES
(22, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzMDAxMTc2LCJpYXQiOjE3NTIzOTYzNzZ9.rCoJGFKVNfHFZAeaGlIdX4a0p5gOApYvWmkep6igpXY', '0000-00-00 00:00:00', '2025-07-13 08:46:16'),
(23, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzMDAxODg5LCJpYXQiOjE3NTIzOTcwODl9.cMHaTZv-IOrPEoP1CYWOg5vlYyzWhh4uFn9qSgqGtPU', '0000-00-00 00:00:00', '2025-07-13 08:58:09'),
(24, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzMDA3ODM4LCJpYXQiOjE3NTI0MDMwMzh9.qvsvbm7JDqd_WlUys4q9RtsqCxY1qkzQWQrvkHwEz5c', '0000-00-00 00:00:00', '2025-07-13 10:37:18'),
(25, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzMDE2NzMyLCJpYXQiOjE3NTI0MTE5MzJ9.xC-7AH3H4RLgNbChKKN0ypiS2YU4-Du7CcUYoe4s4sc', '0000-00-00 00:00:00', '2025-07-13 13:05:32'),
(26, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzMDg5MDQyLCJpYXQiOjE3NTI0ODQyNDJ9.pMgINXN5TAqbDtYPnZ0mw0hbwsMcuunX20kbvHaUViQ', '0000-00-00 00:00:00', '2025-07-14 09:10:42'),
(27, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzNTQ2MjA5LCJpYXQiOjE3NTI5NDE0MDl9.PnMt2u_Ni1oYDa8cM94DC8cV64tRrr1LVDVGm4coMOw', '0000-00-00 00:00:00', '2025-07-19 16:10:09'),
(28, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzNTQ4NTc3LCJpYXQiOjE3NTI5NDM3Nzd9.qxXa3m5WdkvcUb-zMGA3C-yxmvZiJdsZkTbNxBSplAM', '0000-00-00 00:00:00', '2025-07-19 16:49:37'),
(29, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzNTUzODYyLCJpYXQiOjE3NTI5NDkwNjJ9.Ey4OgKccmArQXasezD26YrIDD9gE5MgwTX9xiAPS6p8', '0000-00-00 00:00:00', '2025-07-19 18:17:42'),
(30, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzNTU1MjUwLCJpYXQiOjE3NTI5NTA0NTB9.CxhNnP3mieCfSZw2q_txd2rP4TLD3JvmkrVwVCYxdPM', '0000-00-00 00:00:00', '2025-07-19 18:40:50'),
(31, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzNjE0MTcyLCJpYXQiOjE3NTMwMDkzNzJ9.44UbfJ65AwP8r8yO4w0m7uMu-xmgphlc20CkusncOEU', '0000-00-00 00:00:00', '2025-07-20 11:02:52'),
(32, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzUzNjE1MzI1LCJpYXQiOjE3NTMwMTA1MjV9.QjZXNOxPNEgL0VurSJNdhm0AJDHFWA6ViibGpp6_sok', '0000-00-00 00:00:00', '2025-07-20 11:22:05'),
(33, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJlbWFpbCI6ImlsaXlhQGV4YW1wbGUuY29tIiwicm9sZV9pZCI6MiwiZXhwIjoxNzU0ODI4MjM0LCJpYXQiOjE3NTQyMjM0MzR9.pxmFPBgSRFgQE5rj7FIYb64PN0epC6TQJWo3SiePy2E', '0000-00-00 00:00:00', '2025-08-03 12:17:14');

-- --------------------------------------------------------

--
-- Table structure for table `roles`
--

CREATE TABLE `roles` (
  `id` int(11) NOT NULL,
  `name` varchar(50) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `roles`
--

INSERT INTO `roles` (`id`, `name`) VALUES
(1, 'admin'),
(2, 'member');

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` int(11) NOT NULL,
  `full_name` varchar(100) NOT NULL,
  `email` varchar(100) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `role_id` int(11) NOT NULL,
  `created_at` datetime DEFAULT current_timestamp(),
  `updated_at` datetime DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `full_name`, `email`, `password_hash`, `role_id`, `created_at`, `updated_at`) VALUES
(6, 'iliya', 'iliya@example.com', '$2a$10$ty/Hxu0FH/sMcUoEhC38S.b8oCwJdUH0GQ9NkzOVcQV5OEjzrQvEC', 2, '2025-06-29 16:33:34', '2025-06-29 16:33:34'),
(8, 'iliya', 'iliyaa@example.com', '$2a$10$iQB3asauMDkkC9TTS8HnLe/BeTFsMZxQwsl0Z8qwsmiNXLCNirTIy', 2, '2025-07-01 21:45:44', '2025-07-01 21:45:44'),
(9, 'iliya', 'iliyaaa@example.com', '$2a$10$3E.l8aSyvckdAj3LN3mIV.6FJVa2W4YbHw5SYSRfEQ6P/HvhYzwTO', 2, '2025-07-02 15:03:21', '2025-07-02 15:03:21'),
(11, 'iliya', 'iliyaaaa@example.com', '$2a$10$T31tznUtXLs3wcqSDdlNkujnpyCQGhuCHvHvdeFMc9lAx5rzlm7Ga', 2, '2025-07-13 12:14:18', '2025-07-13 12:14:18');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `authors`
--
ALTER TABLE `authors`
  ADD PRIMARY KEY (`id`);
ALTER TABLE `authors` ADD FULLTEXT KEY `idx_authors_fulltext` (`name`,`biography`);
ALTER TABLE `authors` ADD FULLTEXT KEY `name` (`name`,`biography`);

--
-- Indexes for table `books`
--
ALTER TABLE `books`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `isbn` (`isbn`),
  ADD KEY `author_id` (`author_id`),
  ADD KEY `category_id` (`category_id`);
ALTER TABLE `books` ADD FULLTEXT KEY `title` (`title`,`description`);

--
-- Indexes for table `book_reviews`
--
ALTER TABLE `book_reviews`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_id` (`user_id`),
  ADD KEY `book_id` (`book_id`);

--
-- Indexes for table `categories`
--
ALTER TABLE `categories`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`);

--
-- Indexes for table `loans`
--
ALTER TABLE `loans`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_id` (`user_id`),
  ADD KEY `book_id` (`book_id`);

--
-- Indexes for table `refresh_tokens`
--
ALTER TABLE `refresh_tokens`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `roles`
--
ALTER TABLE `roles`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `email` (`email`),
  ADD KEY `role_id` (`role_id`);
ALTER TABLE `users` ADD FULLTEXT KEY `full_name` (`full_name`,`email`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `authors`
--
ALTER TABLE `authors`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=22;

--
-- AUTO_INCREMENT for table `books`
--
ALTER TABLE `books`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=22;

--
-- AUTO_INCREMENT for table `book_reviews`
--
ALTER TABLE `book_reviews`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `categories`
--
ALTER TABLE `categories`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `loans`
--
ALTER TABLE `loans`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=8;

--
-- AUTO_INCREMENT for table `refresh_tokens`
--
ALTER TABLE `refresh_tokens`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=34;

--
-- AUTO_INCREMENT for table `roles`
--
ALTER TABLE `roles`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=12;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `books`
--
ALTER TABLE `books`
  ADD CONSTRAINT `books_ibfk_1` FOREIGN KEY (`author_id`) REFERENCES `authors` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `books_ibfk_2` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL;

--
-- Constraints for table `book_reviews`
--
ALTER TABLE `book_reviews`
  ADD CONSTRAINT `book_reviews_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `book_reviews_ibfk_2` FOREIGN KEY (`book_id`) REFERENCES `books` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `loans`
--
ALTER TABLE `loans`
  ADD CONSTRAINT `loans_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `loans_ibfk_2` FOREIGN KEY (`book_id`) REFERENCES `books` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `refresh_tokens`
--
ALTER TABLE `refresh_tokens`
  ADD CONSTRAINT `refresh_tokens_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `users`
--
ALTER TABLE `users`
  ADD CONSTRAINT `users_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
